<script setup lang="ts">
import { computed, reactive, ref, onBeforeUnmount } from 'vue';
import { useRouter } from 'vue-router';
import { useStreamStore } from '@/stores/stream';
import { ArrowLeft, Clock, Download, Maximize } from 'lucide-vue-next';

const router = useRouter();

const streamStore = useStreamStore();

streamStore.loadRecordings();

const props = defineProps<{
  recording: string,
}>();

const recording = computed(() => streamStore.recordings.find(r => r.id === props.recording));

const recordingDuration = computed(() => {
  if (!recording.value) {
    return 0;
  }
  return Math.round(((new Date(recording.value.end)).getTime() - (new Date(recording.value.start)).getTime()) / 1000)
});

const data = reactive({
  sliderPos: 0,
  sliderPosInterval: null as ReturnType<typeof setInterval>|null,
});

const video = ref<HTMLVideoElement|null>(null);

const setVideoPos = () => {
  if (!video.value) {
    return;
  }
  video.value.currentTime = data.sliderPos;
};

const formatTime = (seconds: number) => {
  let base = 0;
  if (recording.value) {
    base = (new Date(recording.value.start)).getTime() / 1000;
  }

  const date = new Date((base + seconds) * 1000);
  return date.toLocaleTimeString();
};

data.sliderPosInterval = setInterval(() => {
  if (video.value) {
    data.sliderPos = video.value.currentTime;
  }
}, 100);
onBeforeUnmount(() => {
  if (data.sliderPosInterval !== null) {
    clearInterval(data.sliderPosInterval);
  }
});

const videoContainer = ref<HTMLDivElement | null>(null);

const toggleFullscreen = () => {
  if (!videoContainer.value) return;

  if (!document.fullscreenElement) {
    videoContainer.value.requestFullscreen();
  } else {
    document.exitFullscreen();
  }
};
</script>

<template>
  <div class="h-full flex flex-col bg-gray-900">
    <div class="absolute top-4 left-4 z-10 flex gap-2">
      <button
        @click="router.push({ name: 'recordings' })"
        class="bg-gray-900/80 hover:bg-gray-900 text-white p-3 rounded-md cursor-pointer transition-colors backdrop-blur-sm"
        title="Back to Recordings"
      >
        <ArrowLeft :size="20" />
      </button>
    </div>

    <div class="absolute top-4 right-4 z-10 flex gap-2">
      <button
        v-if="recording"
        @click="router.push({ name: 'camera-timeline', params: { streamId: recording.stream_id } })"
        class="bg-gray-900/80 hover:bg-gray-900 text-white p-3 rounded-md cursor-pointer transition-colors backdrop-blur-sm"
        title="Camera Timeline"
      >
        <Clock :size="20" />
      </button>
      <a
        v-if="recording"
        :href="recording.path"
        download
        class="bg-gray-900/80 hover:bg-gray-900 text-white p-3 rounded-md cursor-pointer transition-colors backdrop-blur-sm"
        title="Download Recording"
      >
        <Download :size="20" />
      </a>
      <button
        @click="toggleFullscreen"
        class="bg-gray-900/80 hover:bg-gray-900 text-white p-3 rounded-md cursor-pointer transition-colors backdrop-blur-sm"
        title="Toggle Fullscreen"
      >
        <Maximize :size="20" />
      </button>
    </div>

    <div v-if="recording" class="flex-1 flex flex-col items-center justify-center relative">
      <div ref="videoContainer" class="w-full h-full max-w-7xl bg-black flex flex-col relative">
        <div class="flex-1 relative min-h-0">
          <video
            ref="video"
            class="w-full h-full"
            :src="recording.path"
            muted
            controls
            autoplay
            playsinline
          />

          <div class="absolute bottom-16 left-4 z-20 bg-gray-900/90 text-white px-4 py-2 rounded-md backdrop-blur-sm max-w-[calc(100%-2rem)]">
            <h2 class="text-lg font-medium mb-1">{{ recording.stream_name }}</h2>
            <div class="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-2 text-sm text-gray-300">
              <RouterLink :to="`/cameras/${recording.stream_id}`" class="hover:underline">
                View Camera
              </RouterLink>
              <span class="hidden sm:inline">•</span>
              <span>{{ (new Date(recording.start)).toLocaleString() }}</span>
              <span class="hidden sm:inline">•</span>
              <span>{{ Math.round(((new Date(recording.end)).getTime() - (new Date(recording.start)).getTime()) / (1000 * 60)) }} min</span>
            </div>
          </div>
        </div>

        <div v-if="recording && recording.performed_motion_detect" class="bg-black text-white px-6 py-4">
          <div class="timeline-container">
            <div class="motion-markers">
              <div
                v-for="m in recording.motion"
                :key="m.t"
                class="motion-marker"
                :style="{
                  left: `${(m.t / recordingDuration) * 100}%`,
                  height: `${Math.max(10, (m.s / 100) * 100)}%`
                }"
                :title="`Motion score: ${m.s}`"
              ></div>
            </div>

            <div class="timeline-track">
              <div class="timeline-progress" :style="{ width: `${(data.sliderPos / recordingDuration) * 100}%` }"></div>
            </div>

            <input
              class="timeline-slider"
              type="range"
              step="any"
              :min="0"
              :max="recordingDuration"
              v-model="data.sliderPos"
              @input="setVideoPos"
            />

            <div class="timeline-labels">
              <div class="timeline-label timeline-label--first">{{ formatTime(0) }}</div>
              <div class="timeline-label timeline-label--mid">{{ formatTime(recordingDuration / 4) }}</div>
              <div class="timeline-label timeline-label--mid">{{ formatTime(recordingDuration / 2) }}</div>
              <div class="timeline-label timeline-label--mid">{{ formatTime(3 * recordingDuration / 4) }}</div>
              <div class="timeline-label timeline-label--last">{{ formatTime(recordingDuration) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="flex-1 flex items-center justify-center text-white">
      <p>Recording not found</p>
    </div>
  </div>
</template>

<style scoped>
.timeline-container {
  position: relative;
  width: 100%;
  height: 60px;
}

.motion-markers {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 20px;
  pointer-events: none;
}

.motion-marker {
  position: absolute;
  bottom: 0;
  width: 2px;
  background-color: #ef4444;
  transform: translateX(-50%);
  opacity: 0.8;
}

.timeline-track {
  position: absolute;
  top: 20px;
  left: 0;
  right: 0;
  height: 4px;
  background-color: #374151;
  border-radius: 2px;
  overflow: hidden;
}

.timeline-progress {
  height: 100%;
  background-color: #6b7280;
  transition: width 0.1s linear;
}

.timeline-slider {
  position: absolute;
  top: 12px;
  left: 0;
  right: 0;
  width: 100%;
  height: 20px;
  -webkit-appearance: none;
  appearance: none;
  background: transparent;
  cursor: pointer;
  outline: none;
  z-index: 10;
}

.timeline-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: white;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  transition: transform 0.1s ease;
}

.timeline-slider::-webkit-slider-thumb:hover {
  transform: scale(1.2);
}

.timeline-slider::-moz-range-thumb {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: white;
  cursor: pointer;
  border: none;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  transition: transform 0.1s ease;
}

.timeline-slider::-moz-range-thumb:hover {
  transform: scale(1.2);
}

.timeline-labels {
  position: absolute;
  top: 32px;
  left: 0;
  right: 0;
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #9ca3af;
  pointer-events: none;
}

.timeline-label {
  user-select: none;
}

@media (max-width: 600px) {
  .timeline-label--mid {
    display: none;
  }
}
</style>

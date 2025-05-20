<script setup lang="ts">
import { computed, reactive, ref, onBeforeUnmount } from 'vue';
import { useRouter } from 'vue-router';
import { useStreamStore } from '@/stores/stream';
import { ArrowLeft, Download, Maximize } from 'lucide-vue-next';

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
      <div ref="videoContainer" class="w-full h-full max-w-7xl bg-black relative">
        <video
          class="w-full h-full"
          :src="recording.path"
          muted
          controls
          autoplay
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

        <div v-if="recording && recording.performed_motion_detect" class="bg-gray-800 text-white p-4">
          <input class="range" type="range" list="range" step="any" :min="0" :max="recordingDuration" v-model="data.sliderPos" @input="setVideoPos" />
          <datalist id="range">
            <option v-for="m in recording.motion" :value="m.t" :label="`${m.s}`"></option>
          </datalist>
        </div>
      </div>
    </div>

    <div v-else class="flex-1 flex items-center justify-center text-white">
      <p>Recording not found</p>
    </div>
  </div>
</template>

<style scoped>
.range {
  width: 100%;
}
</style>
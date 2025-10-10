<script setup lang="ts">
import { computed, reactive, ref, watch, onBeforeUnmount } from 'vue';
import { useRouter } from 'vue-router';
import { useStreamStore } from '@/stores/stream';
import * as types from '@/stores/streamTypes';
import { ArrowLeft, ChevronLeft, ChevronRight, Download, Maximize } from 'lucide-vue-next';

const router = useRouter();
const streamStore = useStreamStore();

streamStore.loadStreams();
streamStore.loadRecordings();

const props = defineProps<{
  streamId: string,
}>();

const stream = computed(() => streamStore.streams.find(s => s.id === props.streamId));

// Filter recordings for this camera, sorted chronologically (oldest first for timeline)
const cameraRecordings = computed(() => {
  return streamStore.recordings
    .filter(r => r.stream_id === props.streamId)
    .sort((a, b) => new Date(a.start).getTime() - new Date(b.start).getTime());
});

// Time window configuration (in milliseconds)
const TIME_WINDOW_OPTIONS = [
  { label: '3 hours', duration: 3 * 60 * 60 * 1000 },
  { label: '6 hours', duration: 6 * 60 * 60 * 1000 },
  { label: '12 hours', duration: 12 * 60 * 60 * 1000 },
  { label: '24 hours', duration: 24 * 60 * 60 * 1000 },
];

const data = reactive({
  timeWindowIndex: 0, // Index into TIME_WINDOW_OPTIONS
  windowEndTime: Date.now(), // End of the time window (defaults to now)
  selectedRecordingId: null as string | null,
  sliderPos: 0,
  sliderPosInterval: null as ReturnType<typeof setInterval> | null,
  seekToPositionOnLoad: null as number | null, // Position to seek to when recording loads
});

const timeWindow = computed(() => TIME_WINDOW_OPTIONS[data.timeWindowIndex]);
const windowStartTime = computed(() => data.windowEndTime - timeWindow.value.duration);

// Recordings visible in current time window
const visibleRecordings = computed(() => {
  return cameraRecordings.value.filter(r => {
    const recStart = new Date(r.start).getTime();
    const recEnd = new Date(r.end).getTime();
    // Recording overlaps with window if it starts before window ends AND ends after window starts
    return recStart < data.windowEndTime && recEnd > windowStartTime.value;
  });
});

// Currently selected recording
const selectedRecording = computed(() => {
  if (!data.selectedRecordingId) return null;
  return cameraRecordings.value.find(r => r.id === data.selectedRecordingId);
});

const recordingDuration = computed(() => {
  if (!selectedRecording.value) return 0;
  return Math.round(
    (new Date(selectedRecording.value.end).getTime() - new Date(selectedRecording.value.start).getTime()) / 1000
  );
});

// Initialize with most recent recording
if (cameraRecordings.value.length > 0) {
  const mostRecent = cameraRecordings.value[cameraRecordings.value.length - 1];
  data.selectedRecordingId = mostRecent.id;
}

const video = ref<HTMLVideoElement | null>(null);

// Update slider position from video
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

const setVideoPos = () => {
  if (!video.value) return;
  video.value.currentTime = data.sliderPos;
};

const formatTime = (timestamp: number) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString();
};

const formatDate = (timestamp: number) => {
  const date = new Date(timestamp);
  return date.toLocaleDateString();
};

// Navigate time window
const goToPreviousWindow = () => {
  data.windowEndTime -= timeWindow.value.duration;
};

const goToNextWindow = () => {
  data.windowEndTime += timeWindow.value.duration;
};

const canGoNext = computed(() => {
  return data.windowEndTime < Date.now();
});

// Handle timeline click to select recording at that time
const handleTimelineClick = (event: MouseEvent) => {
  const target = event.currentTarget as HTMLElement;
  const rect = target.getBoundingClientRect();
  const clickX = event.clientX - rect.left;
  const percentClicked = clickX / rect.width;
  let clickedTime = windowStartTime.value + (timeWindow.value.duration * percentClicked);

  // Snap to nearest motion event if close enough
  const SNAP_THRESHOLD_MS = 60000; // 60 seconds in milliseconds
  let nearestMotionTime: number | null = null;
  let nearestDistance = Infinity;

  // Check all motion events in visible recordings
  visibleRecordings.value.forEach(recording => {
    if (!recording.performed_motion_detect || !recording.motion.length) return;

    const recStart = new Date(recording.start).getTime();
    recording.motion.forEach(motion => {
      const motionTime = recStart + (motion.t * 1000);
      const distance = Math.abs(motionTime - clickedTime);

      if (distance < SNAP_THRESHOLD_MS && distance < nearestDistance) {
        nearestDistance = distance;
        nearestMotionTime = motionTime;
      }
    });
  });

  // If we found a nearby motion event, snap to it
  if (nearestMotionTime !== null) {
    clickedTime = nearestMotionTime;
  }

  // Find recording that contains this time
  const recordingAtTime = visibleRecordings.value.find(r => {
    const recStart = new Date(r.start).getTime();
    const recEnd = new Date(r.end).getTime();
    return clickedTime >= recStart && clickedTime <= recEnd;
  });

  if (recordingAtTime) {
    // Calculate position within the recording
    const recStart = new Date(recordingAtTime.start).getTime();
    const positionInRecording = (clickedTime - recStart) / 1000; // Convert to seconds

    // Check if we're clicking on the same recording or a different one
    if (recordingAtTime.id === data.selectedRecordingId) {
      // Same recording - just seek directly
      data.sliderPos = positionInRecording;
      setVideoPos();
    } else {
      // Different recording - store position and change recording
      data.seekToPositionOnLoad = positionInRecording;
      data.selectedRecordingId = recordingAtTime.id;
    }
  }
};

// Calculate position and width for a recording in the timeline
const getRecordingStyle = (recording: types.Recording) => {
  const recStart = new Date(recording.start).getTime();
  const recEnd = new Date(recording.end).getTime();

  const visibleStart = Math.max(recStart, windowStartTime.value);
  const visibleEnd = Math.min(recEnd, data.windowEndTime);

  const left = ((visibleStart - windowStartTime.value) / timeWindow.value.duration) * 100;
  const width = ((visibleEnd - visibleStart) / timeWindow.value.duration) * 100;

  return {
    left: `${left}%`,
    width: `${width}%`,
  };
};

// Calculate position for motion markers within the timeline
const getMotionMarkerStyle = (recording: types.Recording, motion: types.Motion) => {
  const recStart = new Date(recording.start).getTime();
  const motionTime = recStart + (motion.t * 1000); // motion.t is in seconds

  if (motionTime < windowStartTime.value || motionTime > data.windowEndTime) {
    return { display: 'none' };
  }

  const left = ((motionTime - windowStartTime.value) / timeWindow.value.duration) * 100;
  const height = Math.max(10, (motion.s / 100) * 100);

  return {
    left: `${left}%`,
    height: `${height}%`,
  };
};

// Calculate current playback position in timeline
const currentPlaybackPosition = computed(() => {
  if (!selectedRecording.value) return 0;

  const recStart = new Date(selectedRecording.value.start).getTime();
  const currentTime = recStart + (data.sliderPos * 1000);

  if (currentTime < windowStartTime.value || currentTime > data.windowEndTime) {
    return -1; // Not visible in current window
  }

  return ((currentTime - windowStartTime.value) / timeWindow.value.duration) * 100;
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

// Generate time labels for the timeline
const timeLabels = computed(() => {
  const labels = [];
  const numLabels = 5;

  for (let i = 0; i < numLabels; i++) {
    const time = windowStartTime.value + (timeWindow.value.duration * i / (numLabels - 1));
    labels.push({
      position: (i / (numLabels - 1)) * 100,
      time: formatTime(time),
    });
  }

  return labels;
});

// Handle video loaded metadata - seek to position if needed
const handleVideoLoadedMetadata = () => {
  if (data.seekToPositionOnLoad !== null && video.value) {
    const targetPosition = data.seekToPositionOnLoad;
    data.seekToPositionOnLoad = null; // Clear immediately
    video.value.currentTime = targetPosition;
    data.sliderPos = targetPosition;
  }
};

// Watch for recording changes
watch(() => data.selectedRecordingId, () => {
  // Reset slider position when recording changes (unless we have a specific position to seek to)
  if (data.seekToPositionOnLoad === null) {
    data.sliderPos = 0;
  }
});

// Handle video end - auto-play next recording
const handleVideoEnded = () => {
  if (!selectedRecording.value) return;

  // Find the current recording's index in cameraRecordings
  const currentIndex = cameraRecordings.value.findIndex(r => r.id === selectedRecording.value!.id);

  // If there's a next recording, play it
  if (currentIndex >= 0 && currentIndex < cameraRecordings.value.length - 1) {
    const nextRecording = cameraRecordings.value[currentIndex + 1];
    data.selectedRecordingId = nextRecording.id;
    data.sliderPos = 0;
  }
};
</script>

<template>
  <div class="h-full flex flex-col bg-gray-900">
    <!-- Header Controls -->
    <div class="absolute top-4 left-4 z-10 flex gap-2">
      <button
        @click="router.push({ name: 'cameras' })"
        class="bg-gray-900/80 hover:bg-gray-900 text-white p-3 rounded-md cursor-pointer transition-colors backdrop-blur-sm"
        title="Back to Cameras"
      >
        <ArrowLeft :size="20" />
      </button>
    </div>

    <div class="absolute top-4 right-4 z-10 flex gap-2">
      <a
        v-if="selectedRecording"
        :href="selectedRecording.path"
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

    <!-- Main Content -->
    <div class="flex-1 flex flex-col items-center justify-center relative">
      <div ref="videoContainer" class="w-full h-full max-w-7xl bg-black flex flex-col relative">
        <!-- Video Player -->
        <div class="flex-1 relative min-h-0">
          <div v-if="selectedRecording" class="w-full h-full">
            <video
              ref="video"
              class="w-full h-full"
              :src="selectedRecording.path"
              muted
              controls
              autoplay
              @ended="handleVideoEnded"
              @loadedmetadata="handleVideoLoadedMetadata"
            />

            <div class="absolute bottom-16 left-4 z-20 bg-gray-900/90 text-white px-4 py-2 rounded-md backdrop-blur-sm max-w-[calc(100%-2rem)]">
              <h2 class="text-lg font-medium mb-1">{{ stream?.name || 'Camera' }}</h2>
              <div class="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-2 text-sm text-gray-300">
                <span>{{ new Date(selectedRecording.start).toLocaleString() }}</span>
                <span class="hidden sm:inline">•</span>
                <span>{{ Math.round(((new Date(selectedRecording.end)).getTime() - (new Date(selectedRecording.start)).getTime()) / (1000 * 60)) }} min</span>
              </div>
            </div>
          </div>

          <div v-else class="w-full h-full flex items-center justify-center text-gray-400">
            <p>No recording selected</p>
          </div>
        </div>

        <!-- Multi-Recording Timeline -->
        <div class="bg-black text-white px-6 py-4">
          <!-- Time Window Controls -->
          <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between mb-3 gap-3">
            <div class="flex items-center gap-2">
              <button
                @click="goToPreviousWindow"
                class="bg-gray-800 hover:bg-gray-700 text-white p-2 rounded transition-colors"
                title="Previous Time Window"
              >
                <ChevronLeft :size="20" />
              </button>
              <button
                @click="goToNextWindow"
                :disabled="!canGoNext"
                class="bg-gray-800 hover:bg-gray-700 text-white p-2 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                title="Next Time Window"
              >
                <ChevronRight :size="20" />
              </button>
            </div>

            <div v-if="streamStore.streams.length > 0" class="flex flex-wrap items-center gap-2 sm:gap-3 text-sm">
              <select
                v-once
                :value="streamId"
                @change="router.push({ name: 'camera-timeline', params: { streamId: ($event.target as HTMLSelectElement).value } })"
                class="bg-gray-800 text-white px-2 sm:px-3 py-1 rounded text-sm border border-gray-700"
              >
                <option v-once v-for="s in streamStore.streams" :key="s.id" :value="s.id">
                  {{ s.name }}
                </option>
              </select>
              <span class="text-sm text-gray-400 whitespace-nowrap">{{ formatDate(windowStartTime) }}</span>
              <select
                v-model.number="data.timeWindowIndex"
                class="bg-gray-800 text-white px-2 sm:px-3 py-1 rounded text-sm border border-gray-700"
              >
                <option v-once v-for="(option, idx) in TIME_WINDOW_OPTIONS" :key="`tw-${idx}`" :value="idx">
                  {{ option.label }}
                </option>
              </select>
            </div>
          </div>

          <!-- Timeline Visualization -->
          <div class="timeline-container" @click="handleTimelineClick">
            <!-- Recording Segments -->
            <div class="recording-segments">
              <div
                v-for="recording in visibleRecordings"
                :key="recording.id"
                class="recording-segment"
                :class="{ 'recording-segment--selected': recording.id === data.selectedRecordingId }"
                :style="getRecordingStyle(recording)"
                :title="`${new Date(recording.start).toLocaleTimeString()} - ${new Date(recording.end).toLocaleTimeString()}`"
              ></div>
            </div>

            <!-- Motion Markers -->
            <div class="motion-markers">
              <template v-for="recording in visibleRecordings" :key="recording.id">
                <div
                  v-for="motion in recording.motion"
                  :key="`${recording.id}-${motion.t}`"
                  class="motion-marker"
                  :style="getMotionMarkerStyle(recording, motion)"
                  :title="`Motion score: ${motion.s}`"
                ></div>
              </template>
            </div>

            <!-- Timeline Track -->
            <div class="timeline-track"></div>

            <!-- Current Playback Position -->
            <div
              v-if="currentPlaybackPosition >= 0"
              class="playback-indicator"
              :style="{ left: `${currentPlaybackPosition}%` }"
            ></div>

            <!-- Time Labels -->
            <div class="timeline-labels">
              <div
                v-for="(label, idx) in timeLabels"
                :key="idx"
                class="timeline-label"
                :class="{
                  'timeline-label--first': idx === 0,
                  'timeline-label--last': idx === timeLabels.length - 1
                }"
                :style="{ left: `${label.position}%` }"
              >
                {{ label.time }}
              </div>
            </div>
          </div>

          <!-- Individual Recording Timeline (if recording selected and has motion) -->
          <div v-if="selectedRecording && selectedRecording.performed_motion_detect" class="mt-6 pt-4 border-t border-gray-800">
            <div class="text-xs text-gray-400 mb-2">Current Recording Timeline</div>
            <div class="recording-timeline-container">
              <div class="motion-markers">
                <div
                  v-for="m in selectedRecording.motion"
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
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!stream" class="flex-1 flex items-center justify-center text-white">
      <p>Camera not found</p>
    </div>
  </div>
</template>

<style scoped>
.timeline-container {
  position: relative;
  width: 100%;
  height: 80px;
  cursor: pointer;
  user-select: none;
}

.recording-segments {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 30px;
  pointer-events: none;
}

.recording-segment {
  position: absolute;
  top: 5px;
  height: 20px;
  background-color: #4b5563;
  border-radius: 2px;
  transition: background-color 0.2s;
}

.recording-segment--selected {
  background-color: #3b82f6;
}

.motion-markers {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 30px;
  pointer-events: none;
}

.motion-marker {
  position: absolute;
  bottom: 5px;
  width: 2px;
  background-color: #ef4444;
  transform: translateX(-50%);
  opacity: 0.8;
}

.timeline-track {
  position: absolute;
  top: 35px;
  left: 0;
  right: 0;
  height: 4px;
  background-color: #1f2937;
  border-radius: 2px;
}

.playback-indicator {
  position: absolute;
  top: 30px;
  width: 2px;
  height: 14px;
  background-color: #fbbf24;
  transform: translateX(-50%);
  pointer-events: none;
  z-index: 20;
}

.timeline-labels {
  position: absolute;
  top: 45px;
  left: 0;
  right: 0;
  height: 35px;
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #9ca3af;
  pointer-events: none;
}

.timeline-label {
  position: absolute;
  transform: translateX(-50%);
  white-space: nowrap;
  user-select: none;
}

.timeline-label--first {
  transform: translateX(0);
}

.timeline-label--last {
  transform: translateX(-100%);
}

/* Individual recording timeline */
.recording-timeline-container {
  position: relative;
  width: 100%;
  height: 60px;
}

.recording-timeline-container .motion-markers {
  height: 20px;
}

.recording-timeline-container .timeline-track {
  top: 20px;
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
</style>

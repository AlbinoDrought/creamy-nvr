<script setup lang="ts">
import { ref, computed } from 'vue';
import { X, Scissors, Download, Loader2, AlertCircle } from 'lucide-vue-next';
import { useFFmpeg } from '@/composables/useFFmpeg';
import type { Recording } from '@/stores/streamTypes';

const props = defineProps<{
  visible: boolean;
  clipStart: number; // Unix timestamp in milliseconds
  clipEnd: number; // Unix timestamp in milliseconds
  recordings: Recording[]; // Recordings that overlap with the clip range
}>();

const emit = defineEmits<{
  close: [];
}>();

const ffmpeg = useFFmpeg();
const isProcessing = ref(false);
const error = ref<string | null>(null);
const outputBlob = ref<Blob | null>(null);

const clipDuration = computed(() => {
  const durationMs = props.clipEnd - props.clipStart;
  const seconds = Math.floor(durationMs / 1000);
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}m ${remainingSeconds}s`;
});

const estimatedSize = computed(() => {
  const durationSeconds = (props.clipEnd - props.clipStart) / 1000;
  // Rough estimate: ~1MB per minute at typical recording bitrate
  const sizeMB = (durationSeconds / 60) * 1.0;
  if (sizeMB < 1) {
    return `~${(sizeMB * 1024).toFixed(0)} KB`;
  }
  return `~${sizeMB.toFixed(1)} MB`;
});

const formatTimestamp = (timestamp: number) => {
  return new Date(timestamp).toLocaleString();
};

const handleClose = () => {
  if (isProcessing.value) {
    const confirm = window.confirm('Processing is in progress. Are you sure you want to cancel?');
    if (!confirm) return;
  }
  emit('close');
  // Reset state
  outputBlob.value = null;
  error.value = null;
};

const createClip = async () => {
  isProcessing.value = true;
  error.value = null;
  outputBlob.value = null;

  try {
    // Load FFmpeg if not already loaded
    if (!ffmpeg.isLoaded.value) {
      await ffmpeg.loadFFmpeg();
    }

    // Fetch all required recordings
    const videoBlobs: Blob[] = [];
    let accumulatedDuration = 0; // Duration in seconds from start of first recording
    let trimStartInConcatenated: number | undefined;
    let trimDuration: number | undefined;

    for (const recording of props.recordings) {
      const recStart = new Date(recording.start).getTime();
      const recEnd = new Date(recording.end).getTime();
      const recDuration = (recEnd - recStart) / 1000; // seconds

      // Fetch the video file
      const response = await fetch(recording.path);
      if (!response.ok) {
        throw new Error(`Failed to fetch recording: ${recording.path}`);
      }
      const blob = await response.blob();
      videoBlobs.push(blob);

      // Calculate trim points if this is the first recording
      if (videoBlobs.length === 1) {
        // Calculate start offset within the first recording
        const startOffsetMs = props.clipStart - recStart;
        if (startOffsetMs > 0) {
          trimStartInConcatenated = startOffsetMs / 1000; // convert to seconds
        } else {
          trimStartInConcatenated = 0;
        }
      }

      accumulatedDuration += recDuration;
    }

    // Calculate the trim duration
    const totalClipDuration = (props.clipEnd - props.clipStart) / 1000; // seconds
    trimDuration = totalClipDuration;

    let resultBlob: Blob;

    if (videoBlobs.length === 1) {
      // Single recording - just trim it
      resultBlob = await ffmpeg.trimVideo(
        videoBlobs[0],
        trimStartInConcatenated!,
        trimDuration,
        'clip.mp4'
      );
    } else {
      // Multiple recordings - concatenate then trim
      resultBlob = await ffmpeg.concatenateVideos(
        videoBlobs,
        trimStartInConcatenated,
        trimDuration,
        'clip.mp4'
      );
    }

    outputBlob.value = resultBlob;
  } catch (err) {
    console.error('Clip creation failed:', err);
    error.value = err instanceof Error ? err.message : 'Failed to create clip';
  } finally {
    isProcessing.value = false;
  }
};

const downloadClip = () => {
  if (!outputBlob.value) return;

  const filename = `clip_${new Date(props.clipStart).toISOString().replace(/[:.]/g, '-')}_${clipDuration.value.replace(' ', '')}.mp4`;
  ffmpeg.downloadBlob(outputBlob.value, filename);
};

const handleBackdropClick = (event: MouseEvent) => {
  if (event.target === event.currentTarget) {
    handleClose();
  }
};
</script>

<template>
  <div
    v-if="visible"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm"
    @click="handleBackdropClick"
  >
    <div class="bg-gray-900 rounded-lg shadow-xl w-full max-w-lg mx-4 text-white">
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b border-gray-800">
        <div class="flex items-center gap-2">
          <Scissors :size="24" />
          <h2 class="text-xl font-semibold">Create Video Clip</h2>
        </div>
        <button
          @click="handleClose"
          class="text-gray-400 hover:text-white transition-colors"
          :disabled="isProcessing"
        >
          <X :size="24" />
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 space-y-4">
        <!-- Clip info -->
        <div class="space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-400">Start:</span>
            <span class="font-mono">{{ formatTimestamp(clipStart) }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-400">End:</span>
            <span class="font-mono">{{ formatTimestamp(clipEnd) }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-400">Duration:</span>
            <span class="font-mono">{{ clipDuration }}</span>
          </div>
          <!-- <div class="flex justify-between text-sm">
            <span class="text-gray-400">Estimated size:</span>
            <span class="font-mono">{{ estimatedSize }}</span>
          </div> -->
          <div class="flex justify-between text-sm">
            <span class="text-gray-400">Recordings:</span>
            <span class="font-mono">{{ recordings.length }}</span>
          </div>
        </div>

        <!-- Warning for large clips -->
        <div
          v-if="recordings.length > 5 || (clipEnd - clipStart) / 1000 > 600"
          class="flex items-start gap-2 p-3 bg-yellow-900/30 border border-yellow-600/50 rounded text-sm"
        >
          <AlertCircle :size="16" class="mt-0.5 flex-shrink-0" />
          <div>
            <p class="font-medium">Large clip detected</p>
            <p class="text-gray-300 text-xs mt-1">
              This clip is large and may take several minutes to process. Processing happens entirely in your browser.
            </p>
          </div>
        </div>

        <!-- Progress -->
        <div v-if="isProcessing" class="space-y-2">
          <div class="flex items-center gap-2 text-sm text-gray-300">
            <Loader2 :size="16" class="animate-spin" />
            <span>{{ ffmpeg.currentOperation.value || 'Processing...' }}</span>
          </div>
          <div class="w-full bg-gray-800 rounded-full h-2">
            <div
              class="bg-blue-600 h-2 rounded-full transition-all duration-300"
              :style="{ width: `${ffmpeg.progress.value}%` }"
            ></div>
          </div>
          <p class="text-xs text-gray-400">
            {{ Math.round(ffmpeg.progress.value) }}%
          </p>
        </div>

        <!-- Error -->
        <div
          v-if="error"
          class="flex items-start gap-2 p-3 bg-red-900/30 border border-red-600/50 rounded text-sm"
        >
          <AlertCircle :size="16" class="mt-0.5 flex-shrink-0" />
          <div>
            <p class="font-medium">Error</p>
            <p class="text-gray-300 text-xs mt-1">{{ error }}</p>
          </div>
        </div>

        <!-- Success -->
        <div
          v-if="outputBlob && !isProcessing"
          class="flex items-center gap-2 p-3 bg-green-900/30 border border-green-600/50 rounded text-sm"
        >
          <div class="flex-1">
            <p class="font-medium">Clip ready!</p>
            <p class="text-gray-300 text-xs mt-1">
              Size: {{ (outputBlob.size / 1024 / 1024).toFixed(2) }} MB
            </p>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end gap-3 p-6 border-t border-gray-800">
        <button
          @click="handleClose"
          class="px-4 py-2 text-gray-300 hover:text-white transition-colors"
          :disabled="isProcessing"
        >
          {{ outputBlob ? 'Close' : 'Cancel' }}
        </button>
        <button
          v-if="!outputBlob"
          @click="createClip"
          :disabled="isProcessing || recordings.length === 0"
          class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed rounded transition-colors"
        >
          <Scissors v-if="!isProcessing" :size="16" />
          <Loader2 v-else :size="16" class="animate-spin" />
          <span>{{ isProcessing ? 'Creating...' : 'Create Clip' }}</span>
        </button>
        <button
          v-else
          @click="downloadClip"
          class="flex items-center gap-2 px-4 py-2 bg-green-600 hover:bg-green-700 rounded transition-colors"
        >
          <Download :size="16" />
          <span>Download</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Add any additional custom styles if needed */
</style>

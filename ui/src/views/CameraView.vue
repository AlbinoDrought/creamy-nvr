<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useStreamStore } from '@/stores/stream';
import VideoStream from '@/components/VideoStream.vue';
import { ArrowLeft, Maximize, Volume2, VolumeX } from 'lucide-vue-next';

const router = useRouter();
const streamStore = useStreamStore();

streamStore.loadStreams();

const props = defineProps<{
  streamId: string;
}>();

const stream = computed(() => streamStore.streams.find(s => s.id === props.streamId));
const muted = ref(true);
const videoContainer = ref<HTMLDivElement | null>(null);

const toggleMute = () => {
  muted.value = !muted.value;
};

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
        @click="router.push({ name: 'cameras' })"
        class="bg-gray-900/80 hover:bg-gray-900 text-white p-3 rounded-md cursor-pointer transition-colors backdrop-blur-sm"
        title="Back to Cameras"
      >
        <ArrowLeft :size="20" />
      </button>
    </div>

    <div class="absolute top-4 right-4 z-10 flex gap-2">
      <button
        @click="toggleMute"
        class="bg-gray-900/80 hover:bg-gray-900 text-white p-3 rounded-md cursor-pointer transition-colors backdrop-blur-sm"
        :title="muted ? 'Unmute' : 'Mute'"
      >
        <VolumeX v-if="muted" :size="20" />
        <Volume2 v-if="!muted" :size="20" />
      </button>
      <button
        @click="toggleFullscreen"
        class="bg-gray-900/80 hover:bg-gray-900 text-white p-3 rounded-md cursor-pointer transition-colors backdrop-blur-sm"
        title="Toggle Fullscreen"
      >
        <Maximize :size="20" />
      </button>
    </div>

    <div v-if="stream" class="flex-1 flex flex-col items-center justify-center relative">
      <div ref="videoContainer" class="w-full h-full max-w-7xl bg-black relative">
        <VideoStream
          class="w-full h-full"
          :source="stream.source || `/media/${stream.id}/stream/${stream.id}.m3u8`"
          v-model:muted="muted"
        />

        <div class="absolute bottom-16 left-4 z-20 bg-gray-900/90 text-white px-4 py-2 rounded-md backdrop-blur-sm max-w-[calc(100%-2rem)]">
          <h2 class="text-lg font-medium mb-1">{{ stream.name }}</h2>
          <div class="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-2 text-sm text-gray-300">
            <div class="flex items-center gap-2">
              <span :class="['h-2 w-2 rounded-full flex-shrink-0', { 'bg-green-500': stream.active && !stream.in_err, 'bg-yellow-500': stream.active && stream.in_err, 'bg-red-500': !stream.active }]"></span>
              <span>{{ stream.active ? (stream.in_err ? 'Error' : 'Active') : 'Inactive' }}</span>
            </div>
            <span class="hidden sm:inline">•</span>
            <span>Last Recording: {{ (new Date(stream.last_recording)).toLocaleString() }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="flex-1 flex items-center justify-center text-white">
      <p>Stream not found</p>
    </div>
  </div>
</template>

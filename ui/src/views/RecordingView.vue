<script setup lang="ts">
import { computed, reactive, ref, onBeforeUnmount } from 'vue';
import { useRouter } from 'vue-router';
import { useStreamStore } from '@/stores/stream';

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
</script>

<template>
  <div class="h-full flex flex-col">
    <div class="relative h-full flex flex-col">
      <div class="absolute top-4 right-4 z-10">
        <button @click.prevent="router.push({ name: 'recordings' })" class="bg-gray-900 text-white p-2 rounded cursor-pointer">
          ✕
        </button>
      </div>
      
      <div class="flex-1 bg-gray-700 flex items-center justify-center">
        <div class="w-full h-full flex items-center justify-center bg-gray-700">
          <video
            v-if="recording"
            style="max-height: 80vh"
            :src="recording.path"
            muted
            controls
            autoplay
            ref="video"
          />
        </div>
      </div>
      
      <!-- <div class="h-16 bg-gray-800 flex items-center justify-between px-4">
        <div class="flex items-center gap-4">
          <button class="text-white">
            ⏸️
          </button>
          <button class="text-white">
            🔇
          </button>
          <div class="flex items-center text-white gap-1">
            <span>00:22</span>
            <span>/</span>
            <span>04:12</span>
          </div>
          <div class="w-64 h-1 bg-gray-600 rounded-full overflow-hidden">
            <div class="w-1/4 h-full bg-nvrblue"></div>
          </div>
        </div>
        
        <div class="flex items-center gap-2">
          <button class="text-white">
            📺
          </button>
        </div>
      </div> -->

      <div v-if="recording && recording.performed_motion_detect" class="bg-gray-800 text-white p-4">
        <input class="range" type="range" list="range" step="any" :min="0" :max="recordingDuration" v-model="data.sliderPos" @input="setVideoPos" />
        <datalist id="range">
          <option v-for="m in recording.motion" :value="m.t" :label="`${m.s}`"></option>
        </datalist>
      </div>
      
      <div v-if="recording" class="bg-gray-900 text-white p-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-medium">Info</h3>
          <button>▼</button>
        </div>
        
        <div class="grid grid-cols-2 gap-y-2 mt-4">
          <div>Stream</div>
          <div class="text-right">{{ recording.stream_name }}</div>
          
          <div>Date/Time</div>
          <div class="text-right">{{ (new Date(recording.start)).toLocaleString() }}</div>
          
          <div>Duration</div>
          <div class="text-right">{{ Math.round(((new Date(recording.end)).getTime() - (new Date(recording.start)).getTime()) / (1000 * 60)) }} minutes</div>
          
<!--           
          <div>Unlocked</div>
          <div class="text-right flex items-center justify-end">
            <div class="h-6 w-6 rounded bg-gray-700 mr-1"></div>
            <button class="h-6 w-6 rounded bg-gray-700">🔓</button>
          </div> -->
        </div>
      </div>
      
      <div v-if="recording" class="p-4 bg-gray-100 flex justify-end gap-2">
        <a class="nvr-button bg-blue-500 flex items-center gap-2 py-2" :href="recording.path" download>
          ⬇️ DOWNLOAD
        </a>
        <!-- <button class="border border-gray-300 px-4 py-2 rounded flex items-center gap-2 text-sm">
          🗑️ DELETE
        </button> -->
      </div>
    </div>
  </div>
</template>

<style scoped>
.range {
  width: 100%;
}
</style>
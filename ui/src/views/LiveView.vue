<script setup lang="ts">
import { useStreamStore } from '@/stores/stream';
import VideoStream from '@/components/VideoStream.vue';

const streamStore = useStreamStore();
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- <div class="p-3 border-b border-gray-200 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <select 
          class="border border-gray-200 rounded px-3 py-1.5 text-sm"
          value={selectedView}
          onChange={(e) => setSelectedView(e.target.value)}
        >
          <option>All Cameras</option>
          <option>Front View</option>
          <option>Back View</option>
        </select>
        
        <button class="bg-nvrblue text-white px-3 py-1.5 rounded text-sm">
          EDIT VIEW
        </button>
      </div>
      
      <button class="flex items-center gap-1 text-nvrblue">
        <Maximize2 size={16} />
        <span>Fullscreen</span>
      </button>
    </div> -->
    
    <div class="flex-1 grid grid-cols-1 md:grid-cols-3 gap-2 p-2 bg-gray-900 auto-rows-[minmax(250px,1fr)] md:auto-rows-[auto]" style="grid-template-rows: repeat(auto-fill, 32.5vh)">
      <div
        v-for="(stream, i) in streamStore.streams"
        :key="stream.id"
        :class="['bg-gray-700 flex items-center justify-center', { 'md:col-span-2 md:row-span-2': (i % 6) == 0 }]"
      >
        <VideoStream
          class="w-full h-full"
          :source="stream.source || `/media/${stream.id}/stream/${stream.id}.m3u8`"
        />
      </div>
      <!-- <div class="flex items-center justify-center text-gray-400 bg-black">
        Edit view to add a camera
      </div> -->
    </div>
  </div>
</template>
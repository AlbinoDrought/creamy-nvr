<script setup lang="ts">
import { useRouter } from 'vue-router';
import { Video } from 'lucide-vue-next';
import { useStreamStore } from '@/stores/stream';
import * as types from '@/stores/streamTypes';

const streamStore = useStreamStore();
const router = useRouter();

streamStore.loadStreams();

// const data = reactive({
//   searchText: '',
//   activeTab: 'managed',
//   selectedCameras: [],
// });

// const handleSelectAll = (e: Event) => {
//   throw new Error('todo');
// };

// const handleSelectCamera = (e: Event, id: unknown) => {
//   throw new Error('todo 2');
// };

function viewStream(stream: types.Stream) {
  router.push({ name: 'live-view', query: { stream: stream.id } });
}
</script>

<template>
  <div class="h-full flex">
    <!-- <div class="p-4 flex items-center gap-4">
      <div class="flex items-center gap-2">
        <div class="h-7 w-7 rounded-full bg-gray-200 flex items-center justify-center text-gray-600 text-sm">
          5
        </div>
        
        <div class="relative">
          <input 
            type="text" 
            placeholder="Search" 
            class="search-input pl-8" 
            v-model="data.searchText"
          />
          <Search :size="16" class="absolute left-2.5 top-1/2 transform -translate-y-1/2 text-gray-500" />
        </div>
      </div>
      
      <div class="flex items-center border rounded-md overflow-hidden">
        <button 
          :class="['px-3 py-1.5 text-sm', data.activeTab === 'managed' ? 'bg-nvrblue text-white' : 'bg-white text-gray-700']"
          @click="data.activeTab = 'managed'"
        >
          MANAGED
        </button>
        <button 
          :class="['px-3 py-1.5 text-sm', data.activeTab === 'unmanaged' ? 'bg-nvrblue text-white' : 'bg-white text-gray-700']"
          @click="data.activeTab = 'unmanaged'"
        >
          UNMANAGED
        </button>
      </div>
    </div> -->
    
    <div class="flex-1 px-4 pb-4">
      <table class="w-full nvr-table">
        <thead>
          <tr>
            <th class="w-6">
              <!-- <input 
                type="checkbox" 
                class="h-4 w-4"
                @change="handleSelectAll"
                :checked="data.selectedCameras.length === cameraData.length && cameraData.length > 0"
              /> -->
            </th>
            <th class="text-sm">Name</th>
            <!-- <th class="text-sm">HOST</th> -->
            <!-- <th class="text-sm">MAC ADDRESS</th> -->
            <!-- <th class="text-sm">LINK STATE</th> -->
            <th class="text-sm">Last Segment</th>
            <!-- <th class="text-sm">RECORDING TYPE</th> -->
            <th class="w-36 text-center text-sm"></th>
          </tr>
        </thead>
        <tbody>
            <tr v-for="stream in streamStore.streams" :key="stream.id" class="cursor-pointer" @click.prevent="viewStream(stream)">
              <td>
                <!-- <input 
                  type="checkbox" 
                  class="h-4 w-4"
                  :checked="data.selectedCameras.includes(stream.id)"
                  @change="e => handleSelectCamera(e, stream.id)"
                /> -->
              </td>
              <td class="flex items-center gap-2">
                <span :class="['h-4 w-4 rounded-full', { 'bg-green-500': stream.active && !stream.in_err, 'bg-yellow-500': stream.active && stream.in_err, 'bg-red-500': !stream.active }]"></span>
                {{ stream.name }}
              </td>
              <!-- <td>{{ camera.host }}</td> -->
              <!-- <td>{{ camera.macAddress }}</td> -->
              <!-- <td>
                <div class="flex items-center gap-1">
                  <span class="h-4 w-4 text-green-500">📶</span>
                  <span>{{ camera.linkState }}</span>
                </div>
              </td> -->
              <td>{{ (new Date(stream.last_recording)).toLocaleString() }}</td>
              <!-- <td>{{ camera.recordingType }}</td> -->
              <td>
                <button type="button" class="nvr-button flex items-center justify-center gap-1 mx-auto cursor-pointer" @click.prevent="viewStream(stream)">
                  <Video :size="14" />
                  <span>Watch Live</span>
                </button>
              </td>
            </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

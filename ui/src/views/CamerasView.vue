<script setup lang="ts">
import { Video } from 'lucide-vue-next';
import { useStreamStore } from '@/stores/stream';

const streamStore = useStreamStore();

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
    
    <div class="flex-1 px-4 pb-4 overflow-x-auto">
      <table class="w-full nvr-table min-w-[600px]">
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
            <tr v-for="stream in streamStore.streams" :key="stream.id" class="cursor-pointer">
              <td @click.stop>
                <!-- <input
                  type="checkbox"
                  class="h-4 w-4"
                  :checked="data.selectedCameras.includes(stream.id)"
                  @change="e => handleSelectCamera(e, stream.id)"
                /> -->
              </td>
              <td>
                <RouterLink
                  :to="{ name: 'camera', params: { streamId: stream.id } }"
                  class="flex items-center gap-2"
                >
                  <span :class="['h-4 w-4 rounded-full', { 'bg-green-500': stream.active && !stream.in_err, 'bg-yellow-500': stream.active && stream.in_err, 'bg-red-500': !stream.active }]"></span>
                  {{ stream.name }}
                </RouterLink>
              </td>
              <!-- <td>{{ camera.host }}</td> -->
              <!-- <td>{{ camera.macAddress }}</td> -->
              <!-- <td>
                <div class="flex items-center gap-1">
                  <span class="h-4 w-4 text-green-500">📶</span>
                  <span>{{ camera.linkState }}</span>
                </div>
              </td> -->
              <td>
                <RouterLink
                  :to="{ name: 'camera', params: { streamId: stream.id } }"
                  class="block"
                >
                  {{ (new Date(stream.last_recording)).toLocaleString() }}
                </RouterLink>
              </td>
              <!-- <td>{{ camera.recordingType }}</td> -->
              <td>
                <RouterLink
                  :to="{ name: 'camera', params: { streamId: stream.id } }"
                  class="nvr-button flex items-center justify-center gap-1 mx-auto"
                >
                  <Video :size="14" />
                  <span>Watch Live</span>
                </RouterLink>
              </td>
            </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

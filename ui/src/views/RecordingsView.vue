<script setup lang="ts">
import { reactive } from 'vue';
import { Calendar, Search } from 'lucide-vue-next';
import { useStreamStore } from '@/stores/stream';
import * as types from '@/stores/streamTypes';

const streamStore = useStreamStore();

streamStore.loadRecordings();

const data = reactive({
  // fromDate: null as Date|null,
  // fromTime: '',
  // toDate: null as Date|null,
  // toTime: '',
  // restrictToTimeRange: false,

  // selectedCamera: null,

  // selectedRecordings: [],

  viewing: null as null|types.Recording,
});

// const search = () => {
//   throw new Error('todo');
// };

const viewRecording = (recording: types.Recording) => {
  data.viewing = recording;
};

// const handleSelectRecording = (e: Event, r: unknown) => {
//   throw new Error('todo 3');
// };
</script>

<template>
  <div v-if="data.viewing" class="h-full flex flex-col">
    <div class="relative h-full flex flex-col">
      <div class="absolute top-4 right-4 z-10">
        <button @click.prevent="data.viewing = null" class="bg-gray-900 text-white p-2 rounded cursor-pointer">
          ✕
        </button>
      </div>
      
      <div class="flex-1 bg-gray-700 flex items-center justify-center">
        <div class="w-full h-full flex items-center justify-center bg-gray-700">
          <video
            style="max-height: 80vh"
            :src="data.viewing.path"
            muted
            controls
            autoplay
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
      
      <div class="bg-gray-900 text-white p-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-medium">Info</h3>
          <button>▼</button>
        </div>
        
        <div class="grid grid-cols-2 gap-y-2 mt-4">
          <div>Stream</div>
          <div class="text-right">{{ data.viewing.stream_name }}</div>
          
          <div>Date/Time</div>
          <div class="text-right">{{ (new Date(data.viewing.start)).toLocaleString() }}</div>
          
          <div>Duration</div>
          <div class="text-right">{{ Math.round(((new Date(data.viewing.end)).getTime() - (new Date(data.viewing.start)).getTime()) / (1000 * 60)) }} minutes</div>
          
<!--           
          <div>Unlocked</div>
          <div class="text-right flex items-center justify-end">
            <div class="h-6 w-6 rounded bg-gray-700 mr-1"></div>
            <button class="h-6 w-6 rounded bg-gray-700">🔓</button>
          </div> -->
        </div>
      </div>
      
      <div class="p-4 bg-gray-100 flex justify-end gap-2">
        <a class="nvr-button bg-blue-500 flex items-center gap-2 py-2" :href="data.viewing.path" download>
          ⬇️ DOWNLOAD
        </a>
        <!-- <button class="border border-gray-300 px-4 py-2 rounded flex items-center gap-2 text-sm">
          🗑️ DELETE
        </button> -->
      </div>
    </div>
  </div>
  <div v-else class="h-full flex">
    <!-- <div class="w-80 bg-gray-100 border-r border-gray-200 p-4 flex flex-col gap-6">
      <div>
        <h2 class="font-medium mb-3">DATE RANGE</h2>
        
        <div class="flex gap-4 mb-4">
          <button class="p-2 border border-gray-300 bg-white rounded flex flex-col items-center">
            <Calendar :size="18" class="text-blue-500" />
            <span class="text-xs mt-1">Today</span>
          </button>
          <button class="p-2 border border-gray-300 bg-white rounded flex flex-col items-center">
            <Calendar :size="18" class="text-blue-500" />
            <span class="text-xs mt-1">Last Week</span>
          </button>
          <button class="p-2 border border-gray-300 bg-white rounded flex flex-col items-center">
            <Calendar :size="18" class="text-blue-500" />
            <span class="text-xs mt-1">Last Month</span>
          </button>
        </div>
        
        <div class="mb-3">
          <div class="text-sm mb-1">From</div>
          <div class="flex gap-2">
            <input 
              type="date" 
              class="border border-gray-300 p-2 rounded text-sm w-36"
              v-model="data.fromDate"
            />
            <input 
              type="time" 
              class="border border-gray-300 p-2 rounded text-sm w-24"
              v-model="data.fromTime"
            />
          </div>
        </div>
        
        <div class="mb-3">
          <div class="text-sm mb-1">To</div>
          <div class="flex gap-2">
            <input 
              type="date" 
              class="border border-gray-300 p-2 rounded text-sm w-36"
              v-model="data.toDate"
            />
            <input 
              type="time" 
              class="border border-gray-300 p-2 rounded text-sm w-24"
              v-model="data.toTime"
            />
          </div>
        </div>
        
        <div class="mb-4 flex items-center gap-2">
          <input 
            type="checkbox" 
            id="restrictTime"
            v-model="data.restrictToTimeRange"
            class="h-4 w-4 border border-gray-300"
          />
          <label htmlFor="restrictTime" class="text-sm">Restrict To Time Range</label>
        </div>
        
        <button class="w-full bg-blue-500 text-white py-2 rounded text-sm" @click="search">
          SEARCH
        </button>
      </div>
      
      <div>
        <h2 class="font-medium mb-3">CAMERAS</h2>
        
        <div class="mb-3">
          <select 
            class="w-full border border-gray-300 p-2 rounded text-sm"
            v-model="data.selectedCamera"
          >
            <option value="All Cameras">Select Camera</option>
            <option>todo: replace with cameras</option>
            <option value="Back Deck">Back Deck</option>
            <option value="Driveway">Driveway</option>
            <option value="North West Corner">North West Corner</option>
            <option value="Back Corner">Back Corner</option>
            <option value="North East Corner">North East Corner</option>
          </select>
        </div>
        
        <button class="w-full border border-gray-300 bg-gray-800 text-white py-1.5 px-3 rounded text-xs">
          All Cameras
        </button>
      </div>
      
      <div>
        <h2 class="font-medium mb-3">TYPE</h2>
        
        <div class="flex items-center gap-2 mb-2">
          <input type="checkbox" id="motionRec" class="h-4 w-4" />
          <label htmlFor="motionRec" class="text-sm">Motion Recordings</label>
        </div>
        
        <div class="flex items-center gap-2">
          <input type="checkbox" id="fullTimeRec" class="h-4 w-4" defaultChecked />
          <label htmlFor="fullTimeRec" class="text-sm">Full Time Recordings</label>
        </div>
      </div>
      
      <div>
        <h2 class="font-medium mb-3">LOCKED</h2>
        
        <div class="flex items-center gap-2 mb-2">
          <input type="checkbox" id="lockedRec" class="h-4 w-4" />
          <label htmlFor="lockedRec" class="text-sm">Locked Recordings</label>
        </div>
        
        <div class="flex items-center gap-2">
          <input type="checkbox" id="unlockedRec" class="h-4 w-4" defaultChecked />
          <label htmlFor="unlockedRec" class="text-sm">Unlocked Recordings</label>
        </div>
      </div>
    </div> -->
    
    <div class="flex-1">
      <table class="w-full nvr-table">
        <thead>
          <tr>
            <th class="w-6">
              <!-- <input 
                type="checkbox" 
                class="h-4 w-4"
                onChange={handleSelectAll}
                checked={selectedRecordings.length === recordingData.length && recordingData.length > 0}
              /> -->
            </th>
            <th class="text-sm">Stream</th>
            <!-- <th class="text-sm">TYPE</th> -->
            <th class="text-sm">Date</th>
            <th class="text-sm">Duration</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="recording in streamStore.recordings" :key="recording.id" @click="viewRecording(recording)" class="cursor-pointer">
            <td @click.prevent>
              <!-- <input 
                type="checkbox" 
                class="h-4 w-4"
                :checked="data.selectedRecordings.includes(recording.id)"
                @change="e => handleSelectRecording(e, recording)"
              /> -->
            </td>
            <td>
              <div class="flex items-center gap-3">
                <div class="h-12 w-16 bg-gray-700 rounded overflow-hidden flex-shrink-0">
                  <img :src="recording.thumbnail_path" class="w-full h-full object-cover" loading="lazy" />
                </div>
                <span>{{ recording.stream_name }}</span>
              </div>
            </td>
            <!-- <td>{{ recording.type }}</td> -->
            <td>{{ (new Date(recording.start)).toLocaleString() }}</td>
            <td>{{ Math.round(((new Date(recording.end)).getTime() - (new Date(recording.start)).getTime()) / (1000 * 60)) }} minutes</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
<script setup lang="ts">
import { reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useStreamStore } from '@/stores/stream';
import * as types from '@/stores/streamTypes';

const router = useRouter();

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
});

// const search = () => {
//   throw new Error('todo');
// };

const viewRecording = (recording: types.Recording) => {
  router.push({
    name: 'recording',
    params: {
      recording: recording.id,
    }, 
  });
};

// const handleSelectRecording = (e: Event, r: unknown) => {
//   throw new Error('todo 3');
// };
</script>

<template>
  <div class="h-full flex">
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
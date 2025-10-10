<script setup lang="ts">
import { reactive, computed } from 'vue';
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
});

// Calculate motion density score for a recording
const calculateMotionScore = (recording: types.Recording): number => {
  if (!recording.performed_motion_detect || !recording.motion.length) {
    return 0;
  }

  const duration = ((new Date(recording.end)).getTime() - (new Date(recording.start)).getTime()) / (1000 * 60); // minutes
  const avgScore = recording.motion.reduce((sum, m) => sum + m.s, 0) / recording.motion.length;
  const eventsPerMinute = recording.motion.length / duration;

  // Combine frequency and intensity
  const combinedScore = (eventsPerMinute * 2) + (avgScore / 10);

  return combinedScore;
};

// Calculate per-camera baselines and global percentiles
const motionMetrics = computed(() => {
  const byCamera: Record<string, { scores: number[], mean: number, stdDev: number }> = {};
  const allScores: number[] = [];

  streamStore.recordings.forEach(rec => {
    const score = calculateMotionScore(rec);
    if (score === 0) return; // Skip recordings with no motion

    if (!byCamera[rec.stream_id]) {
      byCamera[rec.stream_id] = { scores: [], mean: 0, stdDev: 0 };
    }
    byCamera[rec.stream_id].scores.push(score);
    allScores.push(score);
  });

  // Calculate stats per camera
  Object.values(byCamera).forEach(camera => {
    const mean = camera.scores.reduce((a, b) => a + b, 0) / camera.scores.length;
    const variance = camera.scores.reduce((sum, score) => sum + Math.pow(score - mean, 2), 0) / camera.scores.length;
    camera.mean = mean;
    camera.stdDev = Math.sqrt(variance);
  });

  // Calculate global percentiles as fallback
  allScores.sort((a, b) => a - b);
  const p33 = allScores[Math.floor(allScores.length * 0.33)] || 0;
  const p66 = allScores[Math.floor(allScores.length * 0.66)] || 0;

  return { byCamera, globalPercentiles: { p33, p66 } };
});

// Determine motion level for a recording
const getMotionLevel = (recording: types.Recording): { level: string, display: string, color: string } => {
  if (!recording.performed_motion_detect) {
    return { level: 'unknown', display: '-', color: 'bg-gray-300 text-gray-600' };
  }

  const score = calculateMotionScore(recording);

  if (score === 0) {
    return { level: 'none', display: 'None', color: 'bg-gray-300 text-gray-600' };
  }

  const cameraStats = motionMetrics.value.byCamera[recording.stream_id];
  let level: string;

  // Use per-camera if we have enough data (5+ recordings)
  if (cameraStats && cameraStats.scores.length >= 5) {
    if (score < cameraStats.mean - 0.3 * cameraStats.stdDev) {
      level = 'low';
    } else if (score < cameraStats.mean + 0.5 * cameraStats.stdDev) {
      level = 'medium';
    } else {
      level = 'high';
    }
  } else {
    // Fallback to global percentiles
    if (score < motionMetrics.value.globalPercentiles.p33) {
      level = 'low';
    } else if (score < motionMetrics.value.globalPercentiles.p66) {
      level = 'medium';
    } else {
      level = 'high';
    }
  }

  const levelMap = {
    low: { display: 'Low', color: 'bg-green-200 text-green-800' },
    medium: { display: 'Medium', color: 'bg-yellow-200 text-yellow-800' },
    high: { display: 'High', color: 'bg-red-200 text-red-800' },
  };

  return { level, ...levelMap[level as keyof typeof levelMap] };
};

// const search = () => {
//   throw new Error('todo');
// };

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
    
    <div class="flex-1 overflow-x-auto">
      <!-- Desktop table view -->
      <table class="w-full nvr-table min-w-[700px] hidden md:table">
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
            <th class="text-sm">Motion</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="recording in streamStore.recordings" :key="recording.id">
            <td @click.stop>
              <!-- <input
                type="checkbox"
                class="h-4 w-4"
                :checked="data.selectedRecordings.includes(recording.id)"
                @change="e => handleSelectRecording(e, recording)"
              /> -->
            </td>
            <td class="!p-0">
              <RouterLink
                :to="{ name: 'recording', params: { recording: recording.id } }"
                class="flex items-center gap-3 px-4 py-3"
              >
                <div class="h-12 w-16 bg-gray-700 rounded overflow-hidden flex-shrink-0">
                  <img :src="recording.thumbnail_path" class="w-full h-full object-cover" loading="lazy" />
                </div>
                <span>{{ recording.stream_name }}</span>
              </RouterLink>
            </td>
            <!-- <td>{{ recording.type }}</td> -->
            <td class="!p-0">
              <RouterLink
                :to="{ name: 'recording', params: { recording: recording.id } }"
                class="block px-4 py-3"
              >
                {{ (new Date(recording.start)).toLocaleString() }}
              </RouterLink>
            </td>
            <td class="!p-0">
              <RouterLink
                :to="{ name: 'recording', params: { recording: recording.id } }"
                class="block px-4 py-3"
              >
                {{ Math.round(((new Date(recording.end)).getTime() - (new Date(recording.start)).getTime()) / (1000 * 60)) }} minutes
              </RouterLink>
            </td>
            <td>
              <span
                :class="['inline-block px-2 py-1 rounded-full text-xs font-medium', getMotionLevel(recording).color]"
              >
                {{ getMotionLevel(recording).display }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Mobile card view -->
      <div class="md:hidden space-y-3 p-4 pt-[60px]">
        <RouterLink
          v-for="recording in streamStore.recordings"
          :key="recording.id"
          :to="{ name: 'recording', params: { recording: recording.id } }"
          class="block bg-white rounded-lg border border-gray-200 overflow-hidden hover:bg-gray-50 transition-colors"
        >
          <div class="flex gap-3 p-3">
            <div class="h-16 w-20 bg-gray-700 rounded overflow-hidden flex-shrink-0">
              <img :src="recording.thumbnail_path" class="w-full h-full object-cover" loading="lazy" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="font-medium text-gray-900 mb-1">{{ recording.stream_name }}</div>
              <div class="text-sm text-gray-600 mb-2">
                {{ (new Date(recording.start)).toLocaleDateString() }}
                {{ (new Date(recording.start)).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }}
              </div>
              <div class="flex items-center gap-2 text-xs text-gray-500">
                <span>{{ Math.round(((new Date(recording.end)).getTime() - (new Date(recording.start)).getTime()) / (1000 * 60)) }} min</span>
                <span class="text-gray-400">•</span>
                <span
                  :class="['inline-block px-2 py-0.5 rounded-full font-medium', getMotionLevel(recording).color]"
                >
                  {{ getMotionLevel(recording).display }}
                </span>
              </div>
            </div>
          </div>
        </RouterLink>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { Camera, Map, Video, ListVideo, Clock, AlertTriangle, Users, Settings } from 'lucide-vue-next';
import { useStreamStore } from '@/stores/stream';
import { computed } from 'vue';

const version = import.meta.env.VERSION || 'Development';
const streamStore = useStreamStore();

const firstCameraId = computed(() => {
  return streamStore.streams.length > 0 ? streamStore.streams[0].id : '';
});

const props = defineProps<{
  mobileOpen?: boolean;
}>();

const emit = defineEmits<{
  close: [];
}>();

const handleLinkClick = () => {
  emit('close');
};
</script>

<template>
  <aside :class="[
    'w-[144px] bg-sidebar min-h-screen flex flex-col overflow-hidden bg-nvrdark text-white sidebar transition-transform duration-300',
    'md:translate-x-0 md:relative',
    'fixed inset-y-0 left-0 z-40',
    mobileOpen ? 'translate-x-0' : '-translate-x-full'
  ]">
    <div class="flex flex-col h-full">
      <div class="flex-1 flex flex-col gap-1 p-2">
        <RouterLink
          to="/cameras"
          class="sidebar-link"
          @click="handleLinkClick"
        >
          <Camera :size="20" />
          <span class="text-xs uppercase font-medium">Cameras</span>
        </RouterLink>

        <!-- <RouterLink
          to="/map"
          class="sidebar-link"
        >
          <Map :size="20" />
          <span class="text-xs uppercase font-medium">Map</span>
        </RouterLink> -->

        <RouterLink
          to="/live-view"
          class="sidebar-link"
          @click="handleLinkClick"
        >
          <Video :size="20" />
          <span class="text-xs uppercase font-medium">Live View</span>
        </RouterLink>

        <RouterLink
          :to="{ name: 'camera-timeline', params: { streamId: firstCameraId || '-' } }"
          class="sidebar-link"
          @click="handleLinkClick"
        >
          <Clock :size="20" />
          <span class="text-xs uppercase font-medium">Timeline</span>
        </RouterLink>

        <RouterLink
          to="/recordings"
          class="sidebar-link"
          @click="handleLinkClick"
        >
          <ListVideo :size="20" />
          <span class="text-xs uppercase font-medium">Recordings</span>
        </RouterLink>
      </div>
      
      <div class="p-3 flex items-center justify-center text-xs text-gray-400">
        {{ version }}
      </div>
      
      <!-- <div class="p-3 flex flex-col gap-1">
        <RouterLink 
          to="/alerts" 
          class="sidebar-link"
        >
          <AlertTriangle :size="20" />
          <span class="text-xs uppercase font-medium">Alerts</span>
        </RouterLink>
        
        <RouterLink 
          to="/users" 
          class="sidebar-link"
        >
          <Users :size="20" />
          <span class="text-xs uppercase font-medium">Users</span>
        </RouterLink>
        
        <RouterLink 
          to="/settings" 
          class="sidebar-link"
        >
          <Settings :size="20" />
          <span class="text-xs uppercase font-medium">Settings</span>
        </RouterLink>
      </div> -->
    </div>
  </aside>
</template>

<script setup lang="ts">
import { Camera, Map, Video, ListVideo, Clock, AlertTriangle, Users, Settings } from 'lucide-vue-next';
const version = import.meta.env.VERSION || 'Development';
</script>

<template>
  <aside class="w-[144px] bg-sidebar min-h-screen flex flex-col overflow-hidden bg-nvrdark text-white sidebar">
    <div class="flex flex-col h-full">
      <div class="flex-1 flex flex-col gap-1 p-2">
        <RouterLink 
          to="/cameras" 
          class="sidebar-link"
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
        >
          <Video :size="20" />
          <span class="text-xs uppercase font-medium">Live View</span>
        </RouterLink>
        
        <!-- <div class="relative">
          <RouterLink 
            to="/timeline" 
            class="sidebar-link"
          >
            <Clock :size="20" />
            <span class="text-xs uppercase font-medium">Timeline</span>
          </RouterLink>
          <div class="absolute top-0 right-2 bg-yellow-500 text-black px-1.5 rounded-sm text-[10px] font-bold">
            BETA
          </div>
        </div> -->
        
        <RouterLink 
          to="/recordings" 
          class="sidebar-link"
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

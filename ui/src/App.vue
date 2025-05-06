<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import Header from '@/components/Header.vue';
import Sidebar from '@/components/Sidebar.vue';
import { useStreamStore } from './stores/stream';

const streamStore = useStreamStore();

streamStore.loadStreams();
streamStore.loadRecordings();
</script>

<template>
  <div className="flex flex-col h-screen">
    <Header />
    <div className="flex flex-1 overflow-hidden">
      <Sidebar />
      <main className="flex-1 overflow-auto">
        <RouterView />
      </main>
    </div>
  </div>
</template>

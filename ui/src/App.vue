<script setup lang="ts">
import { ref, computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { Menu, X } from 'lucide-vue-next'
import Header from '@/components/Header.vue';
import Sidebar from '@/components/Sidebar.vue';
import { useStreamStore } from './stores/stream';

const streamStore = useStreamStore();
const mobileMenuOpen = ref(false);
const route = useRoute();

streamStore.loadStreams();
streamStore.loadRecordings();

// Hide sidebar on full-screen views (camera and recording)
const showSidebar = computed(() => route.name !== 'camera' && route.name !== 'recording');

const toggleMobileMenu = () => {
  mobileMenuOpen.value = !mobileMenuOpen.value;
};

const closeMobileMenu = () => {
  mobileMenuOpen.value = false;
};
</script>

<template>
  <div class="flex flex-col h-screen">
    <Header v-if="showSidebar" />

    <!-- Mobile menu button -->
    <button
      v-if="showSidebar"
      @click="toggleMobileMenu"
      class="md:hidden fixed top-2 left-2 z-50 bg-nvrdark text-white p-2 rounded-md"
    >
      <Menu v-if="!mobileMenuOpen" :size="24" />
      <X v-if="mobileMenuOpen" :size="24" />
    </button>

    <!-- Mobile overlay -->
    <div
      v-if="mobileMenuOpen && showSidebar"
      @click="closeMobileMenu"
      class="md:hidden fixed inset-0 bg-black bg-opacity-50 z-30"
    ></div>

    <div class="flex flex-1 overflow-hidden">
      <Sidebar v-if="showSidebar" :mobile-open="mobileMenuOpen" @close="closeMobileMenu" />
      <main class="flex-1 overflow-auto">
        <RouterView />
      </main>
    </div>
  </div>
</template>

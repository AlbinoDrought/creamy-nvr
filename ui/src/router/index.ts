import { createRouter, createWebHistory } from 'vue-router'
import CamerasView from '@/views/CamerasView.vue'
import CameraView from '@/views/CameraView.vue'
import CameraTimelineView from '@/views/CameraTimelineView.vue'
import RecordingsView from '@/views/RecordingsView.vue';
import RecordingView from '@/views/RecordingView.vue';
import LiveView from '@/views/LiveView.vue';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/cameras',
    },
    {
      path: '/cameras',
      name: 'cameras',
      component: CamerasView,
    },
    {
      path: '/cameras/:streamId',
      name: 'camera',
      component: CameraView,
      props: true,
    },
    {
      path: '/cameras/:streamId/timeline',
      name: 'camera-timeline',
      component: CameraTimelineView,
      props: true,
    },
    {
      path: '/recordings',
      name: 'recordings',
      component: RecordingsView,
    },
    {
      path: '/recordings/:recording',
      name: 'recording',
      component: RecordingView,
      props: true,
    },
    {
      path: '/live-view',
      name: 'live-view',
      component: LiveView,
    },
  ],
})

export default router

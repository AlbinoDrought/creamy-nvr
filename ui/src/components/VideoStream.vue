<script setup lang="ts">
import Hls from 'hls.js';
import { onMounted, onUnmounted, ref } from 'vue';

const video = ref(null);

const emit = defineEmits();
const props = defineProps<{
  source: string;
}>();

onMounted(() => {
  const elVideo = video.value as HTMLVideoElement | null;
  if (!elVideo) {
    throw new Error('Video element ref not found!');
  }
  if (Hls.isSupported()) {
    const hls = new Hls();
    hls.on(Hls.Events.ERROR, (_, data) => {
      emit('error', data);
    });
    hls.attachMedia(elVideo);
    hls.on(Hls.Events.MEDIA_ATTACHED, () => {
      hls.loadSource(props.source);
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        elVideo.play();
      });
    });
    onUnmounted(() => {
      hls.destroy();
    });
  } else if (elVideo.canPlayType('application/vnd.apple.mpegurl')) {
    elVideo.src = props.source;
    elVideo.addEventListener('loadedmetadata',function() {
      elVideo.play();
    });
  }
});
</script>

<template>
  <video ref="video" muted preload="none" />
</template>

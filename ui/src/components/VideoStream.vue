<script setup lang="ts">
import Hls from 'hls.js';
import { onMounted, onUnmounted, ref, watch } from 'vue';

const video = ref(null);

const emit = defineEmits<{
  error: [data: any];
  'update:muted': [muted: boolean];
}>();

const props = withDefaults(defineProps<{
  source: string;
  muted?: boolean;
}>(), {
  muted: true,
});

onMounted(() => {
  const elVideo = video.value as HTMLVideoElement | null;
  if (!elVideo) {
    throw new Error('Video element ref not found!');
  }

  // Listen for volume changes from native controls
  const onVolumeChange = () => {
    emit('update:muted', elVideo.muted);
  };
  elVideo.addEventListener('volumechange', onVolumeChange);

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
      elVideo.removeEventListener('volumechange', onVolumeChange);
    });
  } else if (elVideo.canPlayType('application/vnd.apple.mpegurl')) {
    elVideo.src = props.source;
    const onLoadedMetadata = () => {
      elVideo.play();
    };
    elVideo.addEventListener('loadedmetadata', onLoadedMetadata);
    onUnmounted(() => {
      elVideo.removeEventListener('loadedmetadata', onLoadedMetadata);
      elVideo.removeEventListener('volumechange', onVolumeChange);
    });
  }
});

// Watch muted prop and update video element
watch(() => props.muted, (newMuted) => {
  const elVideo = video.value as HTMLVideoElement | null;
  if (elVideo) {
    elVideo.muted = newMuted;
  }
}, { immediate: true });
</script>

<template>
  <video ref="video" :muted="muted" preload="none" controls playsinline />
</template>

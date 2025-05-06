import { ref } from 'vue'
import { defineStore } from 'pinia'
import * as types from './streamTypes'

const sleep = (ms: number) => {
  return new Promise((resolve) => setTimeout(resolve, ms));
};

export const useStreamStore = defineStore('stream', () => {
  const streams = ref([] as types.Stream[]);
  const recordings = ref([] as types.Recording[]);

  async function loadStreams() {
    const resp = await fetch('/api/streams');
    const json = await resp.json();
    streams.value = json;
  };

  async function loadRecordings() {
    const resp = await fetch('/api/recordings');
    const json = await resp.json();
    recordings.value = json;
  };

  return {
    streams,
    loadStreams,
    recordings,
    loadRecordings,
  };
})

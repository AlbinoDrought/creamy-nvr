import { FFmpeg } from '@ffmpeg/ffmpeg';
import { toBlobURL, fetchFile } from '@ffmpeg/util';
import { ref, readonly } from 'vue';

interface ProgressEvent {
  progress: number;
  time: number;
}

// Singleton FFmpeg instance
let ffmpegInstance: FFmpeg | null = null;
let loadingPromise: Promise<void> | null = null;

const isLoaded = ref(false);
const isLoading = ref(false);
const loadError = ref<string | null>(null);
const progress = ref(0);
const currentOperation = ref<string | null>(null);

export function useFFmpeg() {
  const loadFFmpeg = async () => {
    if (isLoaded.value) return;
    if (loadingPromise) return loadingPromise;

    isLoading.value = true;
    loadError.value = null;

    loadingPromise = (async () => {
      try {
        if (!ffmpegInstance) {
          ffmpegInstance = new FFmpeg();

          // Set up logging
          ffmpegInstance.on('log', ({ message }) => {
            console.log('[FFmpeg]', message);
          });

          // Set up progress tracking
          ffmpegInstance.on('progress', ({ progress: prog, time }: ProgressEvent) => {
            progress.value = prog * 100;
            console.log(`[FFmpeg Progress] ${(prog * 100).toFixed(1)}% (time: ${time})`);
          });
        }

        // Load core from local files (served from /public/ffmpeg/)
        // Using single-threaded version (no SharedArrayBuffer required)
        const baseURL = '/ffmpeg';
        const coreURL = await toBlobURL(`${baseURL}/ffmpeg-core.js`, 'text/javascript');
        const wasmURL = await toBlobURL(`${baseURL}/ffmpeg-core.wasm`, 'application/wasm');

        await ffmpegInstance.load({
          coreURL,
          wasmURL,
        });

        isLoaded.value = true;
        console.log('[FFmpeg] Loaded successfully');
      } catch (error) {
        loadError.value = error instanceof Error ? error.message : 'Failed to load FFmpeg';
        console.error('[FFmpeg] Load error:', error);
        throw error;
      } finally {
        isLoading.value = false;
        loadingPromise = null;
      }
    })();

    return loadingPromise;
  };

  /**
   * Trim a single video file
   * @param videoBlob The video file to trim
   * @param startSeconds Start time in seconds
   * @param durationSeconds Duration in seconds
   * @param filename Output filename
   * @returns Blob of the trimmed video
   */
  const trimVideo = async (
    videoBlob: Blob,
    startSeconds: number,
    durationSeconds: number,
    filename = 'output.mp4'
  ): Promise<Blob> => {
    await loadFFmpeg();
    if (!ffmpegInstance) throw new Error('FFmpeg not initialized');

    currentOperation.value = 'Trimming video...';
    progress.value = 0;

    try {
      const inputName = 'input.mp4';

      // Write input file to FFmpeg's virtual filesystem
      await ffmpegInstance.writeFile(inputName, await fetchFile(videoBlob));

      // Use stream copy for fast trimming (no re-encoding)
      // -ss: start time, -t: duration, -c copy: stream copy (no re-encode)
      await ffmpegInstance.exec([
        '-ss', startSeconds.toString(),
        '-i', inputName,
        '-t', durationSeconds.toString(),
        '-c', 'copy',
        '-avoid_negative_ts', 'make_zero',
        filename
      ]);

      // Read output file
      const data = await ffmpegInstance.readFile(filename);

      // Clean up
      await ffmpegInstance.deleteFile(inputName);
      await ffmpegInstance.deleteFile(filename);

      currentOperation.value = null;
      progress.value = 0;

      return new Blob([data], { type: 'video/mp4' });
    } catch (error) {
      currentOperation.value = null;
      progress.value = 0;
      console.error('[FFmpeg] Trim error:', error);
      throw error;
    }
  };

  /**
   * Concatenate multiple video files and optionally trim the result
   * @param videoBlobs Array of video blobs to concatenate
   * @param trimStart Optional start time in seconds (relative to concatenated video)
   * @param trimDuration Optional duration in seconds
   * @param filename Output filename
   * @returns Blob of the concatenated (and optionally trimmed) video
   */
  const concatenateVideos = async (
    videoBlobs: Blob[],
    trimStart?: number,
    trimDuration?: number,
    filename = 'output.mp4'
  ): Promise<Blob> => {
    await loadFFmpeg();
    if (!ffmpegInstance) throw new Error('FFmpeg not initialized');

    currentOperation.value = 'Concatenating videos...';
    progress.value = 0;

    try {
      // Write all input files
      const inputFiles: string[] = [];
      for (let i = 0; i < videoBlobs.length; i++) {
        const inputName = `input${i}.mp4`;
        await ffmpegInstance.writeFile(inputName, await fetchFile(videoBlobs[i]));
        inputFiles.push(inputName);
      }

      // Create concat list file
      const concatList = inputFiles.map(f => `file '${f}'`).join('\n');
      await ffmpegInstance.writeFile('concat.txt', concatList);

      const concatOutput = trimStart !== undefined ? 'temp_concat.mp4' : filename;

      // Concatenate using concat demuxer
      // This works well when all videos have the same codec/format
      await ffmpegInstance.exec([
        '-f', 'concat',
        '-safe', '0',
        '-i', 'concat.txt',
        '-c', 'copy',
        concatOutput
      ]);

      // If trimming is requested, trim the concatenated result
      if (trimStart !== undefined && trimDuration !== undefined) {
        currentOperation.value = 'Trimming concatenated video...';
        await ffmpegInstance.exec([
          '-ss', trimStart.toString(),
          '-i', concatOutput,
          '-t', trimDuration.toString(),
          '-c', 'copy',
          '-avoid_negative_ts', 'make_zero',
          filename
        ]);
        await ffmpegInstance.deleteFile(concatOutput);
      }

      // Read output file
      const data = await ffmpegInstance.readFile(filename);

      // Clean up all files
      for (const inputFile of inputFiles) {
        await ffmpegInstance.deleteFile(inputFile);
      }
      await ffmpegInstance.deleteFile('concat.txt');
      await ffmpegInstance.deleteFile(filename);

      currentOperation.value = null;
      progress.value = 0;

      return new Blob([data], { type: 'video/mp4' });
    } catch (error) {
      currentOperation.value = null;
      progress.value = 0;
      console.error('[FFmpeg] Concatenate error:', error);
      throw error;
    }
  };

  /**
   * Download a blob as a file
   */
  const downloadBlob = (blob: Blob, filename: string) => {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return {
    loadFFmpeg,
    trimVideo,
    concatenateVideos,
    downloadBlob,
    isLoaded: readonly(isLoaded),
    isLoading: readonly(isLoading),
    loadError: readonly(loadError),
    progress: readonly(progress),
    currentOperation: readonly(currentOperation),
  };
}

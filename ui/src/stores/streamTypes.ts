export interface Stream {
  id: string;
  name: string;
  active: boolean;
  in_err: boolean;
  last_recording: string;
  source: string;
}

export interface Motion {
  /**
   * Time of the motion event in this recording, in seconds
   * @example 72 1m12s
   */
  t: number;
  /**
   * The score of the motion event, 0-100
   * @example 2 low score
   * @example 8 high score
   * @example 16 extreme score
   */
  s: number;
}

export interface Recording {
  id: string;
  stream_id: string;
  stream_name: string;
  start: string;
  end: string;
  path: string;
  thumbnail_path: string;
  /**
   * If false, motion detection hasn't been performed, and .motion will be empty.
   * If true, motion detection has been performed: if .motion is still empty, assume no motion.
   */
  performed_motion_detect: boolean;
  motion: Motion[];
}

export interface Stream {
  id: string;
  name: string;
  active: boolean;
  in_err: boolean;
  last_recording: string;
  source: string;
}

export interface Recording {
  id: string;
  stream_id: string;
  stream_name: string;
  start: string;
  end: string;
  path: string;
  thumbnail_path: string;
}

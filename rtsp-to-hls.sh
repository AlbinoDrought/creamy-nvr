#!/bin/sh

# https://github.com/w23/zenki
# "WTFPL for the rest of this repo (or public domain if you're boring)."

set -e

[ -z RTSP_SOURCE ] && { echo "Usage: RTSP_SOURCE=rtsp://... RTSP_NAME=cctv $0"; exit 1; }

# our camera stream
SOURCE="$RTSP_SOURCE"
# some reasonably-unique name
NAME=${RTSP_NAME:-cctv}

# ffmpeg doesn't autocreate directories
mkdir -p "media/$NAME/stream/segments" # stream directory is 100% temporary and can be removed for cleaning
mkdir -p "media/$NAME/archive" # archive directory is more permanent, should only remove old files
rm -f "media/$NAME/stream/$NAME.m3u8"

HLS_TIME=${HLS_TIME:-5} # 5 seconds, inherent delay, each streamed chunk will be this long
HLS_LIST_SIZE=${HLS_LIST_SIZE:-360}  # keep this many chunks of the above duration
SEGMENT_TIME=${SEGMENT_TIME:-300} # 5 minutes, collate chunks into archives of this length
SEGMENT_WRAP=${SEGMENT_WRAP:-864} # 3 days, keep this many of the above segments

# rtsp_transport: force TCP, output is terrible for me using UDP
# re: keep input framerate
# movflags frag_keyframe+empty_moov: fragment output, place keyframes throughout instead of only at the end
# g 52: at least one keyframe every 52 frames
# reset_timestamps 1: reset "length" for each segment, otherwise the second segment claims to have a length of 10 mins, the third a length of 15 mins, etc
# -fps_mode passthrough: replaces -vsync 0

ffmpeg -loglevel level \
        -rtsp_transport tcp \
        -re \
        -i "$SOURCE" \
        -f hls \
        -fps_mode passthrough \
        -copyts -vcodec copy -acodec copy \
        -movflags frag_keyframe+empty_moov \
        -g 52 \
        -hls_flags delete_segments+append_list \
        -hls_time $HLS_TIME \
        -hls_list_size $HLS_LIST_SIZE \
        -hls_base_url "segments/" \
        -strftime 1 \
        -hls_segment_filename "media/$NAME/stream/segments/$NAME-%06d-%F-%H-%M-%S.ts" \
        "media/$NAME/stream/$NAME.m3u8" \
        -f segment \
        -segment_time $SEGMENT_TIME \
        -segment_atclocktime 1 \
        -segment_wrap $SEGMENT_WRAP \
        -reset_timestamps 1 \
        -strftime 1 \
        -vcodec copy -acodec copy \
        -map 0 \
        "media/$NAME/archive/$NAME-%F-%H-%M-%S.mp4"

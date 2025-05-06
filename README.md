# creamy-nvr

This is not intended for production use. This application probably does not function correctly.

This is a wrapper around the `rtsp-to-hls.sh` script in https://github.com/w23/zenki

## Usage

Create a config like this:

```json
{
  "inputs": [
    {
      "id": "my-camera",
      "name": "Camera",
      "url": "rtsp://foo:bar@127.0.0.1:554/stream"
    }
  ]
}
```

Give it to creamy-nvr by saving it to `config.json` or via env in `CREAMY_NVR_CONFIG`

Build the project with `make`

Run the project with `./creamy-nvr`

View the project at http://localhost:3000

### Debugging

Enable debug mode to see ffmpeg logs:

```json
{
  "debug": true
}
```

### Pruning

Enable pruning to automatically remove recordings or stream-segments that are too old or are taking too much space:

```json
{
  "prune_interval_minutes": 60,
  "inputs": [
    {
      "id": "my-camera",
      "name": "Camera",
      "url": "rtsp://foo:bar@127.0.0.1:554/stream",
      "recording_age_limit_hours": 24,
      "recording_size_limit_megabytes": 20000,
      "stream_age_limit_hours": 6,
      "stream_size_limit_megabytes": 500
    }
  ]
}
```

### Container

```
docker build -t creamy-nvr .
docker run \
  --rm \
  -p 3000:3000 \
  -v creamy-nvr-media:/media \
  creamy-nvr
```
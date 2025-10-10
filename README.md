# Creamy NVR

<p align="center">
    <img src="./.readme/logo.png" alt="Creamy NVR Cat Logo">
    <p align="center">The creamiest selfhosted NVR</p>
    <p align="center">
        <a href="https://github.com/AlbinoDrought/creamy-nvr/blob/master/LICENSE"><img alt="AGPL-3.0 License" src="https://img.shields.io/github/license/AlbinoDrought/creamy-nvr"></a>
    </p>
</p>

This is not intended for production use. This application is unlikely to function correctly.

This is a wrapper around the `rtsp-to-hls.sh` script in https://github.com/w23/zenki

## Features

- Live camera viewing with live rewind / time shifting
- 24/7 recording
- Automatic pruning of recordings based on per-camera storage size or age limit
- View timeline of a camera's motion events across many recordings and scrub through them

## Screenshots

| Feature             | Desktop                                      | Mobile                                              |
| ------------------- | -------------------------------------------- | --------------------------------------------------- |
| Listing Cameras     | ![](./.readme/Camera_View.png)               | ![](./.readme/Camera_View_Mobile.png)               |
| Viewing Camera      | ![](./.readme/Single_Camera_View.png)        | ![](./.readme/Single_Camera_View_Mobile.png)        |
| Viewing Cameras     | ![](./.readme/Live_View.png)                 | ![](./.readme/Live_View_Mobile.png)                 |
| Listing Recordings  | ![](./.readme/Recording_View.png)            | ![](./.readme/Recording_View_Mobile.png)            |
| Viewing Recording   | ![](./.readme/Single_Recording_View.png)     | ![](./.readme/Single_Recording_View_Mobile.png)     |
| Viewing Timeline    | ![](./.readme/Timeline_View.png)             | ![](./.readme/Timeline_View_Mobile.png)             |

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

Enable debug mode to see ffmpeg logs in stdout:

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

### Tuning a Stream's Motion Detection

If you get too many motion events, change the motion detection minimum score. 
It defaults to 10.
To include every single event, set this to -1.

```json
{
  "inputs": [
    {
      "id": "my-camera",
      "name": "Camera",
      "url": "rtsp://foo:bar@127.0.0.1:554/stream",
      "motion_detection_minimum_score": 10
    }
  ]
}
```

### Tuning Motion Detection Threads

If your recordings aren't having motion detected in a timely manner, try assigning extra motion detection workers.

This value defaults to one worker per input or at least two goroutines - whichever is greater.

```json
{
  "motion_detection_workers": 7
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

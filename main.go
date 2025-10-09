package main

import (
	"bytes"
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

//go:embed rtsp-to-hls.sh
var script []byte

//go:embed ui/dist
var ui embed.FS

type Input struct {
	// ID is a path-safe human-friendly identifier for this stream, like "doorbell" or "front-door"
	ID string `json:"id"`
	// Name is a human-friendly identifier for this stream, like "Doorbell" or "Front Door"
	Name string `json:"name"`
	// URL is the connection string for this stream, like "rtsp://user:pass@host:1234/path/here?parameters=sample"
	URL string `json:"url"`

	// RecordingAgeLimitHours is the amount of hours of recordings to keep.
	// If 0, disabled.
	// If 1, keep recordings from the last 60 minutes.
	// If 2, keep recordings from the last 120 minutes, etc.
	// Pruning may be performed on an interval.
	// Recordings older than this age may be kept until pruning occurs.
	RecordingAgeLimitHours int `json:"recording_age_limit_hours"`

	// RecordingSizeLimitMegabytes is the amount of megabytes of recordings to keep.
	// If 0, disabled.
	// If 1000, keep 1GB of recordings.
	// If 1000000, keep 1TB of recordings.
	// Pruning may be performed on an interval.
	// The total size of all recordings may exceed this limit until pruning occurs.
	// The oldest recordings will be removed first.
	RecordingSizeLimitMegabytes int `json:"recording_size_limit_megabytes"`

	// StreamAgeLimitHours is the amount of hours of stream segments to keep.
	// See RecordingAgeLimitHours
	StreamAgeLimitHours int `json:"stream_age_limit_hours"`

	// StreamSizeLimitMegabytes is the amount of megabytes of stream segments to keep.
	// See RecordingSizeLimitMegabytes
	StreamSizeLimitMegabytes int `json:"stream_size_limit_megabytes"`
}

// RecordingDirectory contains .mp4 files saved from this stream
func (i Input) RecordingDirectory() string {
	return fmt.Sprintf("media/%v/archive", i.ID)
}

// StreamSegmentDirectory contains .ts files saved from this stream
func (i Input) StreamSegmentDirectory() string {
	return fmt.Sprintf("media/%v/stream/segments", i.ID)
}

type Config struct {
	// Debug changes the log level to DebugLevel.
	// Importantly, this causes raw stream-capturing command logs to be outputted
	Debug bool `json:"debug"`
	// PruneIntervalMinutes determines how often pruning runs.
	// If 0, disabled.
	// If 1, run every minute.
	// If 60, run every 60 minutes, etc.
	PruneIntervalMinutes int `json:"prune_interval_minutes"`
	// Inputs is the list of input streams we should record
	Inputs []Input `json:"inputs"`
}

type AValue[T any] struct {
	lock  sync.RWMutex
	inner *T
}

func (v *AValue[T]) Load() T {
	v.lock.RLock()
	defer v.lock.RUnlock()
	if v.inner == nil {
		return *new(T)
	}
	return *v.inner
}

func (v *AValue[T]) Store(val T) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.inner = &val
}

func parseRecordingTime(path string) (time.Time, error) {
	if len(path) < 24 {
		return time.Time{}, fmt.Errorf("path is too short, cannot parse recording time: %v", path)
	}
	// abcd-2025-04-23-21-09-05.mp4
	recordingDateRaw := path[len(path)-23 : len(path)-4]
	return time.Parse("2006-01-02-15-04-05", recordingDateRaw)
}

type Stream struct {
	Input Input
	// Active is set to true after the stream-capturing command is started
	// Active is set to false before running the stream-capturing command, after it fails to start, or after it exits
	Active AValue[bool]
	// LastRestart is set to time.Now() before starting the stream-capturing command
	LastRestart AValue[time.Time]
	// LastFileOpened is set to time.Now() when the stream-capturing command emits a message line containing "Opening" and "for writing"
	LastFileOpened AValue[time.Time]
	// LastSegmentOpened is set to time.Now() when the stream-capturing command emits a message line that matches the exprSegmentWriting pattern (`[segment ...] [info] Opening 'segment-name-here.mp4' for writing`)
	LastSegmentOpened AValue[time.Time]
	// LastSegmentOpenedName is set to a segment name like "camera/doorbell/archive/doorbell-1995-07-17-12-31-59.mp4" when
	// the stream-capturing command emits a message line that matches the exprSegmentWriting pattern (`[segment ...] [info] Opening 'segment-name-here.mp4' for writing`)
	LastSegmentOpenedName AValue[string]
	// OnSegmentClosed is called synchronously with the open time and name of the previous segment,
	// after a new segment is opened and the LastSegmentOpened and LastSegmentOpenedName values have been overwritten
	OnSegmentClosed func(opened time.Time, segment string)
	// LastSegmentClosed is set to time.Now() when a previous segment exists and the stream-capturing command emits a message line that matches the exprSegmentWriting pattern (`[segment ...] [info] Opening 'segment-name-here.mp4' for writing`)
	LastSegmentClosed AValue[time.Time]
	// LastErr is set to the error received when starting or when waiting for the stream-capturing command
	LastErr AValue[error]

	// LastRestartInErr is set on a schedule. If true, it indicates LastRestart is less than five minutes ago (= the stream-capturing command could be crashlooping)
	LastRestartInErr AValue[bool]
	// LastFileOpenedInErr is set on a schedule. If true, it indicates LastFileOpened is more than three minutes ago (= the stream-capturing command could be frozen or not recording)
	LastFileOpenedInErr AValue[bool]
	// LastSegmentOpenedInErr is set on a schedule. If true, it indicates LastSegmentOpened is more than 15 minutes ago (= the stream-capturing command could be frozen or not recording)
	LastSegmentOpenedInErr AValue[bool]

	// RestartRecording can be invoked to stop and restart the rtsp-to-hls.sh process
	RestartRecording AValue[func()]
}

type Recording struct {
	ID      string
	InputID string
	Start   time.Time
	End     time.Time
	Path    string
}

type ApiV1Stream struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Active        bool   `json:"active"`
	InErr         bool   `json:"in_err"`
	LastRecording string `json:"last_recording"`
	Source        string
}

type ApiV1Recording struct {
	ID            string `json:"id"`
	StreamID      string `json:"stream_id"`
	StreamName    string `json:"stream_name"`
	Start         string `json:"start"`
	End           string `json:"end"`
	Path          string `json:"path"`
	ThumbnailPath string `json:"thumbnail_path"`
}

var logger = logrus.New()

func sizeOfDir(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func main() {
	// todo: store segments to tmpfs
	// todo: compress and migrate segments from local to remote storage

	logger.SetFormatter(&logrus.JSONFormatter{})

	var (
		configBytes []byte
		err         error
	)

	env := os.Getenv("CREAMY_NVR_CONFIG")

	if env == "" {
		configBytes, err = os.ReadFile("config.json")
		if err != nil {
			logger.WithError(err).Fatal("failed to read config.json")
		}
	} else {
		configBytes = []byte(env)
	}

	config := Config{}
	if err = json.Unmarshal(configBytes, &config); err != nil {
		logger.WithError(err).WithField("raw-config", string(configBytes)).Fatal("failed to unmarshal config")
	}

	if config.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	if len(config.Inputs) == 0 {
		logger.Fatal("must have at least one stream")
	}

	ctx := context.Background()

	if _, err := os.Stat("rtsp-to-hls.sh"); os.IsNotExist(err) {
		logger.Warn("rtsp-to-hls.sh not found, using embedded copy")
		if err := os.WriteFile("rtsp-to-hls.sh", script, 0700); err != nil {
			logger.WithError(err).Error("failed to write rtsp-to-hls.sh, please create it manually")
			os.Exit(1)
		}
	}

	recordings := []Recording{}
	recordingsLock := sync.RWMutex{}
	saveRecording := make(chan Recording)
	sortRecording := make(chan bool)

	removeRecordingFromMem := func(path string) {
		recordingsLock.Lock()
		defer recordingsLock.Unlock()

		for i, r := range recordings {
			if r.Path == path {
				recordings = append(recordings[:i], recordings[i+1:]...)
				break
			}
		}
	}

	pruneLock := sync.Mutex{}
	prune := func() {
		pruneLock.Lock()
		defer pruneLock.Unlock()
		logger := logger.WithField("unit", "prune")

		logger.Debug("performing prune")

		for _, input := range config.Inputs {
			// prune recordings by date
			if input.RecordingAgeLimitHours > 0 {
				target := time.Now().Add(-1 * time.Hour * time.Duration(input.RecordingAgeLimitHours))
				err := filepath.Walk(input.RecordingDirectory(), func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}
					if (!strings.HasSuffix(path, ".mp4") && !strings.HasSuffix(path, ".mp4.jpg")) || !strings.Contains(path, input.ID) || len(path) <= 24 {
						return fmt.Errorf("unexpected file found in recording directory: %v", path)
					}

					recordingDate, err := parseRecordingTime(path)
					if err != nil {
						return fmt.Errorf("failed to parse time from path %v: %v", path, err)
					}

					if recordingDate.After(target) {
						return filepath.SkipAll // we have reduced enough storage already
					}

					if err := os.Remove(path); err != nil {
						return fmt.Errorf("failed pruning recording at %v: %v", path, err)
					}
					logger.WithField("path", path).WithField("input", input.ID).Debug("pruned recording due to date")

					removeRecordingFromMem(path)

					return nil
				})
				if err != nil {
					logger.WithError(err).WithField("input", input.ID).Error("failed to perform recording date prune")
				}
				logger.WithField("target", target).WithField("input", input.ID).Debug("pruned recordings by date")
			}

			// prune recordings by size
			if input.RecordingSizeLimitMegabytes > 0 {
				target := int64(input.RecordingSizeLimitMegabytes) * 1000 * 1000
				size, err := sizeOfDir(input.RecordingDirectory())
				if err != nil {
					logger.WithError(err).WithField("input", input.ID).Error("failed to get size of recording dir")
				} else if size > target {
					newSize := size
					err = filepath.Walk(input.RecordingDirectory(), func(path string, info fs.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if newSize <= target {
							return filepath.SkipAll // we have reduced enough storage already
						}
						if info.IsDir() {
							return nil
						}
						if (!strings.HasSuffix(path, ".mp4") && !strings.HasSuffix(path, ".mp4.jpg")) || !strings.Contains(path, input.ID) {
							return fmt.Errorf("unexpected file found in recording directory: %v", path)
						}
						recordingSize := info.Size()
						if err := os.Remove(path); err != nil {
							return fmt.Errorf("failed pruning recording at %v: %v", path, err)
						}
						logger.WithField("path", path).WithField("input", input.ID).Debug("pruned recording due to size")
						newSize -= recordingSize

						removeRecordingFromMem(path)

						return nil
					})
					if err != nil {
						logger.WithError(err).WithField("input", input.ID).Error("failed to perform recording size limit prune")
					}
					logger.WithField("target", target).WithField("size", size).WithField("new-size", newSize).WithField("input", input.ID).Debug("pruned recordings by size")
				}
			}

			// prune stream segments by date
			if input.StreamAgeLimitHours > 0 {
				target := time.Now().Add(-1 * time.Hour * time.Duration(input.StreamAgeLimitHours))
				err := filepath.Walk(input.StreamSegmentDirectory(), func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}
					if !strings.HasSuffix(path, ".ts") || !strings.Contains(path, input.ID) || len(path) <= 24 {
						return fmt.Errorf("unexpected file found in stream segment directory: %v", path)
					}

					// abcd-000001-2025-05-01-22-21-52.ts
					streamSegmentDateRaw := path[len(path)-22 : len(path)-3]
					streamSegmentDate, err := time.Parse("2006-01-02-15-04-05", streamSegmentDateRaw)
					if err != nil {
						return fmt.Errorf("failed to parse time from path %v: %v", path, err)
					}

					if streamSegmentDate.After(target) {
						return filepath.SkipAll // we have reduced enough storage already
					}

					if err := os.Remove(path); err != nil {
						return fmt.Errorf("failed pruning stream segment at %v: %v", path, err)
					}
					logger.WithField("path", path).WithField("input", input.ID).Debug("pruned stream segment due to date")

					return nil
				})
				if err != nil {
					logger.WithError(err).WithField("input", input.ID).Error("failed to perform stream segment date prune")
				}
				logger.WithField("target", target).WithField("input", input.ID).Debug("pruned stream segments by date")
			}

			// prune stream segments by size
			if input.StreamSizeLimitMegabytes > 0 {
				target := int64(input.StreamSizeLimitMegabytes) * 1000 * 1000
				size, err := sizeOfDir(input.StreamSegmentDirectory())
				if err != nil {
					logger.WithError(err).WithField("input", input.ID).Error("failed to get size of stream segment dir")
				} else if size > target {
					newSize := size
					err = filepath.Walk(input.StreamSegmentDirectory(), func(path string, info fs.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if newSize <= target {
							return filepath.SkipAll // we have reduced enough storage already
						}
						if info.IsDir() {
							return nil
						}
						if !strings.HasSuffix(path, ".ts") || !strings.Contains(path, input.ID) {
							return fmt.Errorf("unexpected file found in stream segment directory: %v", path)
						}
						streamSegmentSize := info.Size()
						if err := os.Remove(path); err != nil {
							return fmt.Errorf("failed pruning stream segment at %v: %v", path, err)
						}
						logger.WithField("path", path).WithField("input", input.ID).Debug("pruned stream segment due to size")
						newSize -= streamSegmentSize
						return nil
					})
					if err != nil {
						logger.WithError(err).WithField("input", input.ID).Error("failed to perform stream segment size limit prune")
					}
					logger.WithField("target", target).WithField("size", size).WithField("new-size", newSize).WithField("input", input.ID).Debug("pruned stream segments by size")
				}
			}
		}
	}

	if config.PruneIntervalMinutes > 0 {
		go func() {
			ticker := time.NewTicker(time.Minute * time.Duration(config.PruneIntervalMinutes))
			defer ticker.Stop()
			for range ticker.C {
				prune()
			}
		}()
	}
	go prune()

	go func() {
		for {
			select {
			case recording := <-saveRecording:
				recordingsLock.Lock()
				recordings = append(recordings, recording)
				logger.WithField("recording", recording).Debug("new recording")
				recordingsLock.Unlock()
			case <-sortRecording:
				recordingsLock.Lock()
				sort.Slice(recordings, func(i, j int) bool {
					if recordings[i].Start.Before(recordings[j].Start) {
						return true
					}
					if recordings[i].Start.After(recordings[j].Start) {
						return false
					}
					return recordings[i].Path < recordings[j].Path
				})
				logger.Debug("sorted recordings")
				recordingsLock.Unlock()
			}
		}
	}()

	makeSaveRecording := func(inputIdx int) func(time.Time, string) {
		// recordingIdx := uint64(0)

		thumbnailQueue := make(chan string, 1)
		go func() {
			loggerInfo := logger.WriterLevel(logrus.DebugLevel)
			genThumbnail := func(segment string) error {
				ctx, cancel := context.WithTimeout(ctx, time.Minute)
				defer cancel()
				cmd := exec.CommandContext(
					ctx,
					"ffmpeg",
					"-i", segment,
					"-vframes", "1",
					"-vf", "scale=256:192:force_original_aspect_ratio=decrease",
					segment+".jpg",
				)
				cmd.Stdout = loggerInfo
				cmd.Stderr = loggerInfo
				return cmd.Run()
			}
			for segment := range thumbnailQueue {
				if err := genThumbnail(segment); err != nil {
					logger.WithField("segment", segment).WithError(err).Warn("failed to generate thumbnail, ignoring")
				}
			}
		}()

		return func(opened time.Time, segment string) {
			// newRecordingIdx := atomic.AddUint64(&recordingIdx, 1)
			saveRecording <- Recording{
				// ID:      fmt.Sprintf("%v_%v-%v", time.Now().Unix(), inputIdx, newRecordingIdx),
				ID:      path.Base(segment),
				InputID: config.Inputs[inputIdx].ID,
				Start:   opened,
				End:     time.Now(),
				Path:    segment,
			}
			thumbnailQueue <- segment
		}
	}

	streams := make([]Stream, len(config.Inputs))
	for i := range config.Inputs {
		streams[i].Input = config.Inputs[i]
		streams[i].Active.Store(false)
		streams[i].LastRestart.Store(time.Date(1995, 7, 17, 0, 1, 2, 3, time.Local))
		streams[i].LastFileOpened.Store(time.Date(1995, 7, 17, 0, 1, 2, 3, time.Local))
		streams[i].LastSegmentOpened.Store(time.Date(1995, 7, 17, 0, 1, 2, 3, time.Local))
		streams[i].LastSegmentOpenedName.Store("")
		streams[i].OnSegmentClosed = makeSaveRecording(i)
		streams[i].LastSegmentClosed.Store(time.Date(1995, 7, 17, 0, 1, 2, 3, time.Local))
		streams[i].LastErr.Store(errors.New("empty"))
		streams[i].LastFileOpenedInErr.Store(true)
		streams[i].LastSegmentOpenedInErr.Store(true)
		streams[i].LastRestartInErr.Store(true)
		go record(ctx, &streams[i])
	}

	streamIdxMap := make(map[string]int, len(streams))
	for i := range streams {
		streamIdxMap[streams[i].Input.ID] = i
	}

	go func() {
		for _, input := range config.Inputs {
			err := filepath.Walk(input.RecordingDirectory(), func(fpath string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !strings.HasSuffix(fpath, ".mp4") || !strings.Contains(fpath, input.ID) {
					return nil
				}

				recordingDate, err := parseRecordingTime(fpath)
				if err != nil {
					return fmt.Errorf("failed to parse time from path %v: %v", fpath, err)
				}

				ctx, cancel := context.WithTimeout(ctx, time.Minute)
				defer cancel()
				ffprobe := exec.CommandContext(
					ctx,
					"ffprobe", fpath,
					"-v", "quiet",
					"-of", "json",
					"-show_entries", "format",
				)
				data, err := ffprobe.Output()
				if err != nil {
					logger.WithError(err).WithField("path", fpath).Warn("failed to run ffprobe on old recording, will not appear in UI, ignoring")
					// partial recordings cause exit code 1. ignore those and keep going.
					return nil
				}
				type ffprobeOutput struct {
					Format struct {
						Duration string `json:"duration"`
					} `json:"format"`
				}
				var parsed ffprobeOutput
				if err := json.Unmarshal(data, &parsed); err != nil {
					return fmt.Errorf("failed to parse ffprobe output %v: %v", string(data), err)
				}
				duration, err := strconv.ParseFloat(parsed.Format.Duration, 64)
				if err != nil {
					return fmt.Errorf("failed to parse ffprobe output duration %v: %v", parsed.Format.Duration, err)
				}
				durationI := int(duration)
				end := recordingDate.Add(time.Second * time.Duration(durationI))

				saveRecording <- Recording{
					ID:      path.Base(fpath),
					InputID: input.ID,
					Start:   recordingDate,
					End:     end,
					Path:    fpath, // todo: make into subdir
				}
				logger.WithField("unit", "recordings-loader").WithField("input", input.ID).Debug("loaded recording")

				return nil
			})
			sortRecording <- true
			if err != nil {
				logger.WithError(err).WithField("input", input.ID).Warn("failed to parse old recordings, ignoring")
			}
		}
	}()

	go func() {
		timer := time.NewTimer(30 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				for i := range streams {
					stream := &streams[i]
					lastRestart := stream.LastRestart.Load()
					lastRestartInErr := time.Since(lastRestart) < 5*time.Minute
					lastRestartCurrentlyInErr := stream.LastRestartInErr.Load()

					if lastRestartInErr && !lastRestartCurrentlyInErr {
						logger.WithField("last-restart", lastRestart).Warn("stream restarted less than 5 minutes ago")
					} else if !lastRestartInErr && lastRestartCurrentlyInErr {
						logger.WithField("last-restart", lastRestart).Info("stream has been up at least 5 minutes")
					}

					lastFileOpened := stream.LastFileOpened.Load()
					lastFileOpenedInErr := time.Since(lastFileOpened) > 3*time.Minute
					lastFileOpenedCurrentlyInErr := stream.LastFileOpenedInErr.Load()

					if lastFileOpenedInErr && !lastFileOpenedCurrentlyInErr {
						logger.WithField("last-open", lastFileOpened).Warn("stream has not opened file for at least 3 minutes, restarting")
						restart := stream.RestartRecording.Load()
						if restart != nil {
							restart()
						} else {
							logger.Warn("stream.RestartRecording is nil!")
						}
					} else if !lastFileOpenedInErr && lastFileOpenedCurrentlyInErr {
						logger.WithField("last-open", lastFileOpened).Info("stream has opened a file in the last 3 minutes")
					}

					lastSegmentOpened := stream.LastSegmentOpened.Load()
					lastSegmentOpenedInErr := time.Since(lastSegmentOpened) > 15*time.Minute
					lastSegmentOpenedCurrentlyInErr := stream.LastSegmentOpenedInErr.Load()

					if lastSegmentOpenedInErr && !lastSegmentOpenedCurrentlyInErr {
						logger.WithField("last-open", lastSegmentOpened).Warn("stream has not opened segment for at least 15 minutes, restarting")
						restart := stream.RestartRecording.Load()
						if restart != nil {
							restart()
						} else {
							logger.Warn("stream.RestartRecording is nil!")
						}
					} else if !lastSegmentOpenedInErr && lastSegmentOpenedCurrentlyInErr {
						logger.WithField("last-open", lastSegmentOpened).Info("stream has opened a segment in the last 15 minutes")
					}

					stream.LastRestartInErr.Store(lastRestartInErr)
					stream.LastFileOpenedInErr.Store(lastFileOpenedInErr)
					stream.LastSegmentOpenedInErr.Store(lastSegmentOpenedInErr)
				}
				timer.Reset(time.Minute)
			}
		}
	}()

	sub, err := fs.Sub(ui, "ui/dist")
	if err != nil {
		logger.WithError(err).Fatal("failed to move into ui/dist subfolder of embedded UI bundle")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/streams", func(w http.ResponseWriter, r *http.Request) {
		apiStreams := make([]ApiV1Stream, len(streams))
		for i := range apiStreams {
			apiStreams[i].ID = streams[i].Input.ID
			apiStreams[i].Name = streams[i].Input.Name
			apiStreams[i].Active = streams[i].Active.Load()
			apiStreams[i].InErr = !apiStreams[i].Active || streams[i].LastFileOpenedInErr.Load() || streams[i].LastSegmentOpenedInErr.Load() || streams[i].LastRestartInErr.Load()
			apiStreams[i].LastRecording = streams[i].LastSegmentClosed.Load().Format(time.RFC3339)
			apiStreams[i].Source = fmt.Sprintf("/media/%v/stream/%v.m3u8", streams[i].Input.ID, streams[i].Input.ID)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&apiStreams)
	})
	mux.HandleFunc("GET /api/recordings", func(w http.ResponseWriter, r *http.Request) {
		recordingsLock.RLock()
		defer recordingsLock.RUnlock()
		apiRecordings := make([]ApiV1Recording, len(recordings))
		for i := range apiRecordings {
			revIdx := len(recordings) - i - 1
			streamIdx := streamIdxMap[recordings[revIdx].InputID]
			apiRecordings[i].ID = recordings[revIdx].ID
			apiRecordings[i].StreamID = recordings[revIdx].InputID
			apiRecordings[i].StreamName = streams[streamIdx].Input.Name
			apiRecordings[i].Start = recordings[revIdx].Start.Format(time.RFC3339)
			apiRecordings[i].End = recordings[revIdx].End.Format(time.RFC3339)
			apiRecordings[i].Path = "/" + strings.TrimPrefix(recordings[revIdx].Path, "/")
			apiRecordings[i].ThumbnailPath = "/" + strings.TrimPrefix(recordings[revIdx].Path+".jpg", "/")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&apiRecordings)
	})
	mux.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))
	mux.Handle("/cameras", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, sub, "index.html")
	}))
	mux.Handle("/live-view", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, sub, "index.html")
	}))
	mux.Handle("/recordings", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, sub, "index.html")
	}))
	mux.Handle("/recordings/{file}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, sub, "index.html")
	}))
	mux.Handle("/", http.FileServerFS(sub))

	if err := http.ListenAndServe(":3000", mux); err != nil {
		logger.WithError(err).Error("http.ListenAndServe error")
	}
	logger.Info("end of main")
}

type OpeningForWritingWriter struct {
	stream     *Stream
	parentErr  io.Writer
	parentWarn io.Writer
	parentInfo io.Writer
	buf        bytes.Buffer
}

var exprSegmentWriting = regexp.MustCompile(`\[segment[^\]]+\] \[info\] Opening '([^']+\.mp4)' for writing`)

func (w *OpeningForWritingWriter) onLine(p []byte) {
	str := string(p)
	// {"level":"debug","msg":"[segment @ 0x55e5928f9400] [info] Opening 'camera/doorbell/archive/doorbell-2025-05-01-20-00-25.mp4' for writing","stream":"doorbell","time":"2025-05-01T20:00:25-07:00"}
	// {"level":"debug","msg":"[hls @ 0x55f7d4140740] [info] Opening 'camera/north_west_corner/stream/north_west_corner.m3u8.tmp' for writing","stream":"north_west_corner","time":"2025-05-01T20:00:30-07:00"}
	if strings.Contains(str, "Opening") && strings.Contains(str, "for writing") {
		w.stream.LastFileOpened.Store(time.Now())

		segment := exprSegmentWriting.FindStringSubmatch(str)
		if segment != nil {
			previousSegmentTime := w.stream.LastSegmentOpened.Load()
			previousSegment := w.stream.LastSegmentOpenedName.Load()
			w.stream.LastSegmentOpened.Store(time.Now())
			w.stream.LastSegmentOpenedName.Store(segment[1])
			if previousSegment != "" {
				if w.stream.OnSegmentClosed != nil {
					w.stream.OnSegmentClosed(previousSegmentTime, previousSegment)
				}
				w.stream.LastSegmentClosed.Store(time.Now())
			}
		}
	}

	if w.parentErr != nil {
		if strings.HasPrefix(str, "[fatal]") || strings.HasPrefix(str, "[error]") {
			w.parentErr.Write(p)
		}
	}

	if w.parentWarn != nil {
		if strings.HasPrefix(str, "[warning]") {
			w.parentWarn.Write(p)
		}
	}

	if w.parentInfo != nil {
		w.parentInfo.Write(p)
	}
}

func (w *OpeningForWritingWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	if bytes.ContainsRune(p, '\n') {
		for {
			line, err := w.buf.ReadBytes('\n')
			if err != nil {
				w.buf.Reset()
				w.buf.Write(line)
				break
			}
			w.onLine(line)
		}
	}
	return len(p), nil
}

func record(ctx context.Context, stream *Stream) {
	logger := logger.WithField("stream", stream.Input.ID)
	loggerErr := logger.WriterLevel(logrus.ErrorLevel)
	loggerWarn := logger.WriterLevel(logrus.WarnLevel)
	loggerInfo := logger.WriterLevel(logrus.DebugLevel)

	newCmd := func() *exec.Cmd {
		cmd := exec.CommandContext(ctx, "./rtsp-to-hls.sh")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
		cmd.Env = append(
			cmd.Env,
			"RTSP_SOURCE="+stream.Input.URL,
			"RTSP_NAME="+stream.Input.ID,
		)
		// ffmpeg is writing all output to stderr for me
		cmd.Stderr = &OpeningForWritingWriter{
			stream:     stream,
			parentErr:  loggerErr,
			parentWarn: loggerWarn,
			parentInfo: loggerInfo,
		}
		var cmdRestartLock sync.Mutex
		stream.RestartRecording.Store(func() {
			cmdRestartLock.Lock()
			defer cmdRestartLock.Unlock()
			if cmd.Process != nil {
				syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				logger.Info("performed kill")
			} else {
				logger.Warn("restart recording was requested, but process is nil")
			}
		})
		return cmd
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			stream.Active.Store(false)
			stream.LastRestart.Store(time.Now())

			cmd := newCmd()
			if err := cmd.Start(); err != nil {
				stream.LastErr.Store(err)
				logger.WithError(err).Error("failed to start cmd")
				continue
			}

			stream.Active.Store(true)
			logger.Info("stream active")
			if err := cmd.Wait(); err != nil {
				stream.LastErr.Store(err)
				logger.WithError(err).Error("cmd stopped with error")
			}
			stream.Active.Store(false)
			logger.Info("stream inactive")
		}

		time.Sleep(30 * time.Second)
	}
}

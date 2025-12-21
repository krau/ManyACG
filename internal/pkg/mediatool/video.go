package mediatool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/krau/ffmpeg-go"
	"github.com/samber/oops"
	"github.com/yapingcat/gomedia/go-mp4"
)

type VideoMetadata struct {
	Duration uint // in milliseconds
	Width    uint
	Height   uint
}

// a go native way to get mp4 video metadata
func GetMP4Meta(rs io.ReadSeeker) (*VideoMetadata, error) {
	d := mp4.CreateMp4Demuxer(rs)

	tracks, err := d.ReadHead()
	if err != nil {
		return nil, err
	}

	for _, track := range tracks {
		if track.Cid == mp4.MP4_CODEC_H264 {
			info := d.GetMp4Info()
			return &VideoMetadata{
				Duration: uint(float64(info.Duration) / float64(info.Timescale) * 1000),
				Width:    uint(track.Width),
				Height:   uint(track.Height),
			}, nil
		}
	}

	return nil, fmt.Errorf("no h264 track found")
}

type ffmpegVideoMetadata struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

// GetVideoMetadata uses ffprobe to get video metadata
func GetVideoMetadata(rs io.ReadSeeker) (*VideoMetadata, error) {
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		defer pipeWriter.Close()
		rs.Seek(0, io.SeekStart)
		io.Copy(pipeWriter, rs)
	}()

	result, err := ffmpeg.ProbeReaderWithTimeout(
		pipeReader,
		time.Second*10,
		ffmpeg.KwArgs{
			"select_streams": "v:0",
			"show_entries":   "stream=width,height:format=duration",
			"of":             "json",
		},
	)
	if err != nil {
		return nil, err
	}

	var data ffmpegVideoMetadata

	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, err
	}

	var durationMs uint
	if data.Format.Duration != "" {
		if durationFloat, err := strconv.ParseFloat(data.Format.Duration, 64); err == nil {
			durationMs = uint(durationFloat * 1000)
		}
	}

	meta := &VideoMetadata{
		Duration: durationMs,
	}

	if len(data.Streams) > 0 {
		meta.Width = uint(data.Streams[0].Width)
		meta.Height = uint(data.Streams[0].Height)
	}

	return meta, nil
}

func ExtractVideoThumbFrame(r io.Reader) ([]byte, error) {
	data, err := extractVideoFrameAt(r, 1.0)
	if err == nil && len(data) > 0 {
		return data, nil
	}
	return extractVideoFrameAt(r, 0.0)
}

func extractVideoFrameAt(r io.Reader, timestamp float64) ([]byte, error) {
	var out bytes.Buffer
	var errBuf bytes.Buffer

	err := ffmpeg.
		Input("pipe:0").
		Output("pipe:1", ffmpeg.KwArgs{
			"ss":      fmt.Sprintf("%.3f", timestamp),
			"vframes": 1,
			"f":       "mjpeg",
		}).
		WithInput(r).
		WithOutput(&out).
		OverWriteOutput().
		WithErrorOutput(&errBuf).
		Run()

	if err != nil {
		return nil, oops.Wrapf(err, "ffmpeg error: %s", errBuf.String())
	}

	return out.Bytes(), nil
}

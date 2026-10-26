// Package audio holds audio helpers: probing now, take scoring and mixing later.
package audio

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ProbeSeconds reads a file's duration with ffprobe, rounded to the nearest second.
func ProbeSeconds(ctx context.Context, path string) (int, error) {
	out, err := exec.CommandContext(ctx, "ffprobe", "-v", "error",
		"-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path).Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe %s: %w", path, err)
	}
	secs, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("ffprobe %s: unexpected output %q", path, out)
	}
	return int(secs + 0.5), nil
}

// ProbeBytesSeconds is ProbeSeconds for audio held in memory.
func ProbeBytesSeconds(ctx context.Context, b []byte, ext string) (int, error) {
	f, err := os.CreateTemp("", "probe-*"+ext)
	if err != nil {
		return 0, err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(b); err != nil {
		f.Close()
		return 0, err
	}
	if err := f.Close(); err != nil {
		return 0, err
	}
	return ProbeSeconds(ctx, f.Name())
}

package logtail

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mvaliolahi/tinker/internal/logfmt"
)

const (
	DefaultTailLines  = 20
	DefaultPollInterval = 250 * time.Millisecond
	ChunkSize         = 8192
)

// Follow opens a file, shows the last N lines formatted, then follows for new lines.
func Follow(path string, tailLines int) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer f.Close()

	lines := ReadLastLines(f, tailLines)
	for _, line := range lines {
		fmt.Println(logfmt.FormatLine(line))
	}

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(DefaultPollInterval)
				continue
			}
			return err
		}
		fmt.Println(logfmt.FormatLine(strings.TrimRight(line, "\n")))
	}
}

// ReadLastLines reads the last N lines from a file without loading the entire file.
func ReadLastLines(f *os.File, n int) []string {
	info, err := f.Stat()
	if err != nil {
		return nil
	}
	size := info.Size()
	if size == 0 {
		return nil
	}

	off := size - ChunkSize
	if off < 0 {
		off = 0
	}

	buf := make([]byte, size-off)
	if _, err := f.ReadAt(buf, off); err != nil {
		return nil
	}

	allLines := strings.Split(strings.TrimRight(string(buf), "\n"), "\n")
	if len(allLines) > n {
		allLines = allLines[len(allLines)-n:]
	}

	return allLines
}

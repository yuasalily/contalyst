package dockerx

import (
	"bufio"
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/yuasalily/contalyst/internal/engine"
)

// LogStream follows a container's logs and returns a channel of lines. The
// caller cancels ctx to stop streaming; the channel is closed when the stream
// ends. Stdout/stderr are demultiplexed correctly based on the container's TTY
// setting (inception R1): non-TTY streams carry an 8-byte frame header and must
// go through stdcopy, TTY streams are raw.
func (c *Client) LogStream(ctx context.Context, id string, timestamps bool) (<-chan engine.LogLine, error) {
	tty, err := c.HasTTY(ctx, id)
	if err != nil {
		return nil, err
	}

	reader, err := c.api.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: timestamps,
		Tail:       "500", // seed with recent history without unbounded memory
	})
	if err != nil {
		return nil, err
	}

	out := make(chan engine.LogLine, 256)

	// source is the demuxed, plain-text byte stream.
	var source io.Reader = reader
	var pw *io.PipeWriter
	if !tty {
		var pr *io.PipeReader
		pr, pw = io.Pipe()
		source = pr
		go func() {
			_, copyErr := stdcopy.StdCopy(pw, pw, reader)
			pw.CloseWithError(copyErr)
		}()
	}

	go func() {
		defer close(out)
		defer func() { _ = reader.Close() }()
		if pw != nil {
			defer func() { _ = pw.Close() }()
		}
		scanner := bufio.NewScanner(source)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case out <- engine.LogLine{Text: scanner.Text()}:
			}
		}
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			select {
			case out <- engine.LogLine{Err: err}:
			case <-ctx.Done():
			}
		}
	}()

	return out, nil
}

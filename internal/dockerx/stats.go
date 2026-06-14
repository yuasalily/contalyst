package dockerx

import (
	"context"
	"encoding/json"

	"github.com/docker/docker/api/types/container"
)

// Stats is a single sampled snapshot of a container's resource usage.
type Stats struct {
	CPUPercent float64
	MemUsage   uint64
	MemLimit   uint64
	MemPercent float64
	NetRx      uint64
	NetTx      uint64
	BlkRead    uint64
	BlkWrite   uint64
	Pids       uint64
	Err        error
}

// StatsStream streams resource usage for a container roughly once per second.
// The caller cancels ctx to stop; the channel closes when the stream ends.
func (c *Client) StatsStream(ctx context.Context, id string) (<-chan Stats, error) {
	resp, err := c.api.ContainerStats(ctx, id, true)
	if err != nil {
		return nil, err
	}

	out := make(chan Stats, 8)
	go func() {
		defer close(out)
		defer func() { _ = resp.Body.Close() }()
		dec := json.NewDecoder(resp.Body)
		for {
			if ctx.Err() != nil {
				return
			}
			var raw container.StatsResponse
			if err := dec.Decode(&raw); err != nil {
				if ctx.Err() == nil {
					select {
					case out <- Stats{Err: err}:
					case <-ctx.Done():
					}
				}
				return
			}
			select {
			case <-ctx.Done():
				return
			case out <- computeStats(raw):
			}
		}
	}()
	return out, nil
}

// computeStats derives human-meaningful figures from a raw stats sample. CPU%
// must be computed from the delta between the current and previous samples
// (inception FR-C11), not read directly.
func computeStats(s container.StatsResponse) Stats {
	out := Stats{
		MemUsage: s.MemoryStats.Usage,
		MemLimit: s.MemoryStats.Limit,
		Pids:     s.PidsStats.Current,
	}

	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemUsage) - float64(s.PreCPUStats.SystemUsage)
	onlineCPUs := float64(s.CPUStats.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = float64(len(s.CPUStats.CPUUsage.PercpuUsage))
	}
	if sysDelta > 0 && cpuDelta > 0 && onlineCPUs > 0 {
		out.CPUPercent = (cpuDelta / sysDelta) * onlineCPUs * 100.0
	}

	// On cgroup v2, "usage" includes page cache; subtract inactive_file to match
	// the figure `docker stats` reports.
	if cache, ok := s.MemoryStats.Stats["inactive_file"]; ok && out.MemUsage >= cache {
		out.MemUsage -= cache
	}
	if out.MemLimit > 0 {
		out.MemPercent = float64(out.MemUsage) / float64(out.MemLimit) * 100.0
	}

	for _, n := range s.Networks {
		out.NetRx += n.RxBytes
		out.NetTx += n.TxBytes
	}
	for _, b := range s.BlkioStats.IoServiceBytesRecursive {
		switch b.Op {
		case "read", "Read":
			out.BlkRead += b.Value
		case "write", "Write":
			out.BlkWrite += b.Value
		}
	}
	return out
}

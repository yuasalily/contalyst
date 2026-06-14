package dockerx

import (
	"math"
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestUnit_PrimaryName(t *testing.T) {
	cases := map[string]struct {
		in   []string
		want string
	}{
		"strips slash": {[]string{"/web"}, "web"},
		"first only":   {[]string{"/a", "/b"}, "a"},
		"empty":        {nil, "<none>"},
	}
	for name, c := range cases {
		if got := primaryName(c.in); got != c.want {
			t.Errorf("%s: primaryName(%v)=%q want %q", name, c.in, got, c.want)
		}
	}
}

func TestUnit_ShortImage(t *testing.T) {
	if got := shortImage("sha256:abcdef0123456789"); got != "abcdef012345" {
		t.Errorf("digest: got %q", got)
	}
	if got := shortImage("nginx:alpine"); got != "nginx:alpine" {
		t.Errorf("normal: got %q", got)
	}
}

func TestUnit_FormatPorts(t *testing.T) {
	ports := []container.Port{
		{PrivatePort: 80, PublicPort: 8088, Type: "tcp"},
		{PrivatePort: 80, PublicPort: 8088, Type: "tcp"}, // dup, should collapse
		{PrivatePort: 9000, PublicPort: 0, Type: "tcp"},  // unpublished, should be skipped
	}
	got := formatPorts(ports)
	want := "8088→80/tcp"
	if got != want {
		t.Errorf("formatPorts=%q want %q", got, want)
	}
	if formatPorts(nil) != "" {
		t.Errorf("nil ports should format to empty")
	}
}

func TestUnit_ComputeStats(t *testing.T) {
	raw := container.StatsResponse{}
	raw.CPUStats.CPUUsage.TotalUsage = 200
	raw.CPUStats.SystemUsage = 2000
	raw.CPUStats.OnlineCPUs = 2
	raw.PreCPUStats.CPUUsage.TotalUsage = 100
	raw.PreCPUStats.SystemUsage = 1000
	raw.MemoryStats.Usage = 100
	raw.MemoryStats.Limit = 200
	raw.MemoryStats.Stats = map[string]uint64{"inactive_file": 40}
	raw.PidsStats.Current = 5
	raw.Networks = map[string]container.NetworkStats{
		"eth0": {RxBytes: 1000, TxBytes: 500},
		"eth1": {RxBytes: 200, TxBytes: 100},
	}
	raw.BlkioStats.IoServiceBytesRecursive = []container.BlkioStatEntry{
		{Op: "read", Value: 4096},
		{Op: "write", Value: 2048},
	}

	s := computeStats(raw)

	// cpuDelta=100, sysDelta=1000, online=2 => (100/1000)*2*100 = 20%
	if math.Abs(s.CPUPercent-20.0) > 0.001 {
		t.Errorf("CPUPercent=%.4f want 20", s.CPUPercent)
	}
	// usage 100 - inactive_file 40 = 60; limit 200 => 30%
	if s.MemUsage != 60 {
		t.Errorf("MemUsage=%d want 60", s.MemUsage)
	}
	if math.Abs(s.MemPercent-30.0) > 0.001 {
		t.Errorf("MemPercent=%.4f want 30", s.MemPercent)
	}
	if s.NetRx != 1200 || s.NetTx != 600 {
		t.Errorf("net rx/tx = %d/%d want 1200/600", s.NetRx, s.NetTx)
	}
	if s.BlkRead != 4096 || s.BlkWrite != 2048 {
		t.Errorf("blk r/w = %d/%d want 4096/2048", s.BlkRead, s.BlkWrite)
	}
	if s.Pids != 5 {
		t.Errorf("Pids=%d want 5", s.Pids)
	}
}

func TestUnit_ComputeStatsNoDeltaIsZero(t *testing.T) {
	// First sample (no previous) must not produce NaN/Inf or a bogus percentage.
	raw := container.StatsResponse{}
	raw.CPUStats.CPUUsage.TotalUsage = 0
	raw.CPUStats.SystemUsage = 0
	s := computeStats(raw)
	if s.CPUPercent != 0 || math.IsNaN(s.CPUPercent) {
		t.Errorf("CPUPercent should be 0 with no delta, got %v", s.CPUPercent)
	}
}

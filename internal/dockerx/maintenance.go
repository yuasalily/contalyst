package dockerx

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"
)

// Layer is one entry in an image's history (U12 / FR-L1): a build step with the
// disk space it added and the command that produced it.
type Layer struct {
	Size      int64
	CreatedBy string
	Created   time.Time
}

// ImageHistory returns an image's layers, newest first (as the daemon reports).
func (c *Client) ImageHistory(ctx context.Context, id string) ([]Layer, error) {
	hist, err := c.api.ImageHistory(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]Layer, 0, len(hist))
	for _, h := range hist {
		out = append(out, Layer{
			Size:      h.Size,
			CreatedBy: h.CreatedBy,
			Created:   time.Unix(h.Created, 0),
		})
	}
	return out, nil
}

// PruneKind identifies a category in the prune dashboard (U12 / FR-PR1).
type PruneKind int

const (
	PruneKindImages PruneKind = iota
	PruneKindContainers
	PruneKindVolumes
	PruneKindNetworks
	PruneKindBuildCache
)

func (k PruneKind) Label() string {
	switch k {
	case PruneKindContainers:
		return "stopped containers"
	case PruneKindVolumes:
		return "volumes"
	case PruneKindNetworks:
		return "networks"
	case PruneKindBuildCache:
		return "build cache"
	default:
		return "images"
	}
}

// Usage is the reclaimable-space summary for one prune category.
type Usage struct {
	Kind        PruneKind
	Count       int
	Reclaimable int64
}

// DiskUsage summarises reclaimable space per category for the prune dashboard
// (FR-PR1). Values are best-effort: where the daemon does not report a size
// (e.g. RefCount/Containers == -1) the item is treated as in-use and excluded.
func (c *Client) DiskUsage(ctx context.Context) ([]Usage, error) {
	du, err := c.api.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		return nil, err
	}

	var imgN int
	var imgSize int64
	for _, im := range du.Images {
		if im.Containers == 0 { // unused (in-use or unknown are >0 / -1)
			imgN++
			imgSize += im.Size
		}
	}

	var ctN int
	var ctSize int64
	for _, ct := range du.Containers {
		if !isUpState(string(ct.State)) {
			ctN++
			ctSize += ct.SizeRw
		}
	}

	var volN int
	var volSize int64
	for _, v := range du.Volumes {
		if v.UsageData != nil && v.UsageData.RefCount == 0 {
			volN++
			if v.UsageData.Size > 0 {
				volSize += v.UsageData.Size
			}
		}
	}

	var bcN int
	var bcSize int64
	for _, b := range du.BuildCache {
		if !b.InUse {
			bcN++
			bcSize += b.Size
		}
	}

	netN := c.removableNetworks(ctx)

	return []Usage{
		{Kind: PruneKindImages, Count: imgN, Reclaimable: imgSize},
		{Kind: PruneKindContainers, Count: ctN, Reclaimable: ctSize},
		{Kind: PruneKindVolumes, Count: volN, Reclaimable: volSize},
		{Kind: PruneKindNetworks, Count: netN, Reclaimable: 0},
		{Kind: PruneKindBuildCache, Count: bcN, Reclaimable: bcSize},
	}, nil
}

// removableNetworks counts user-defined networks (the predefined bridge/host/
// none are never pruned), best-effort.
func (c *Client) removableNetworks(ctx context.Context) int {
	nets, err := c.Networks(ctx)
	if err != nil {
		return 0
	}
	n := 0
	for _, net := range nets {
		switch net.Name {
		case "bridge", "host", "none":
		default:
			n++
		}
	}
	return n
}

// Prune runs the prune for one dashboard category (FR-PR2).
func (c *Client) Prune(ctx context.Context, k PruneKind) error {
	switch k {
	case PruneKindContainers:
		return c.PruneContainers(ctx)
	case PruneKindVolumes:
		return c.PruneVolumes(ctx)
	case PruneKindNetworks:
		return c.PruneNetworks(ctx)
	case PruneKindBuildCache:
		_, err := c.api.BuildCachePrune(ctx, build.CachePruneOptions{All: true})
		return err
	default:
		return c.PruneImages(ctx)
	}
}

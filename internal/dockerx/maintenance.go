package dockerx

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"

	"github.com/yuasalily/contalyst/internal/engine"
)

// ImageHistory returns an image's layers, newest first (as the daemon reports).
func (c *Client) ImageHistory(ctx context.Context, id string) ([]engine.Layer, error) {
	hist, err := c.api.ImageHistory(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]engine.Layer, 0, len(hist))
	for _, h := range hist {
		out = append(out, engine.Layer{
			Size:      h.Size,
			CreatedBy: h.CreatedBy,
			Created:   time.Unix(h.Created, 0),
		})
	}
	return out, nil
}

// DiskUsage summarises reclaimable space per category for the prune dashboard
// (FR-PR1). Values are best-effort: where the daemon does not report a size
// (e.g. RefCount/Containers == -1) the item is treated as in-use and excluded.
func (c *Client) DiskUsage(ctx context.Context) ([]engine.Usage, error) {
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
		if !engine.IsUpState(string(ct.State)) {
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

	return []engine.Usage{
		{Kind: engine.PruneKindImages, Count: imgN, Reclaimable: imgSize},
		{Kind: engine.PruneKindContainers, Count: ctN, Reclaimable: ctSize},
		{Kind: engine.PruneKindVolumes, Count: volN, Reclaimable: volSize},
		{Kind: engine.PruneKindNetworks, Count: netN, Reclaimable: 0},
		{Kind: engine.PruneKindBuildCache, Count: bcN, Reclaimable: bcSize},
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
func (c *Client) Prune(ctx context.Context, k engine.PruneKind) error {
	switch k {
	case engine.PruneKindContainers:
		return c.PruneContainers(ctx)
	case engine.PruneKindVolumes:
		return c.PruneVolumes(ctx)
	case engine.PruneKindNetworks:
		return c.PruneNetworks(ctx)
	case engine.PruneKindBuildCache:
		_, err := c.api.BuildCachePrune(ctx, build.CachePruneOptions{All: true})
		return err
	default:
		return c.PruneImages(ctx)
	}
}

package dockerx

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
)

// Image is the UI-facing view model for an image.
type Image struct {
	ID      string
	Repo    string
	Tag     string
	Size    int64
	Created time.Time
}

func (c *Client) Images(ctx context.Context) ([]Image, error) {
	list, err := c.api.ImageList(ctx, image.ListOptions{All: false})
	if err != nil {
		return nil, err
	}
	out := make([]Image, 0, len(list))
	for _, im := range list {
		repo, tag := "<none>", "<none>"
		if len(im.RepoTags) > 0 {
			if r, t, ok := strings.Cut(im.RepoTags[0], ":"); ok {
				repo, tag = r, t
			}
		}
		out = append(out, Image{
			ID:      strings.TrimPrefix(im.ID, "sha256:")[:12],
			Repo:    repo,
			Tag:     tag,
			Size:    im.Size,
			Created: time.Unix(im.Created, 0),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
	return out, nil
}

func (c *Client) RemoveImage(ctx context.Context, id string, force bool) error {
	_, err := c.api.ImageRemove(ctx, id, image.RemoveOptions{Force: force, PruneChildren: true})
	return err
}

func (c *Client) PruneImages(ctx context.Context) error {
	_, err := c.api.ImagesPrune(ctx, filters.Args{})
	return err
}

// Volume is the UI-facing view model for a volume.
type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
}

func (c *Client) Volumes(ctx context.Context) ([]Volume, error) {
	resp, err := c.api.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]Volume, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		out = append(out, Volume{Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (c *Client) RemoveVolume(ctx context.Context, name string, force bool) error {
	return c.api.VolumeRemove(ctx, name, force)
}

func (c *Client) PruneVolumes(ctx context.Context) error {
	_, err := c.api.VolumesPrune(ctx, filters.Args{})
	return err
}

// Network is the UI-facing view model for a network.
type Network struct {
	ID     string
	Name   string
	Driver string
	Scope  string
}

func (c *Client) Networks(ctx context.Context) ([]Network, error) {
	list, err := c.api.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]Network, 0, len(list))
	for _, n := range list {
		id := n.ID
		if len(id) > 12 {
			id = id[:12]
		}
		out = append(out, Network{ID: id, Name: n.Name, Driver: n.Driver, Scope: n.Scope})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (c *Client) RemoveNetwork(ctx context.Context, id string) error {
	return c.api.NetworkRemove(ctx, id)
}

func (c *Client) PruneNetworks(ctx context.Context) error {
	_, err := c.api.NetworksPrune(ctx, filters.Args{})
	return err
}

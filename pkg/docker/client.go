package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// Client wraps the Docker client
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.cli.Close()
}

// PortMapping represents a port mapping
type PortMapping struct {
	IP          string
	PrivatePort uint16
	PublicPort  uint16
	Type        string
}

// ContainerInfo contains container information
type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	Command string
	Created string
	Status  string
	State   string
	Health  string
	Ports   []PortMapping
	Size    string
}

// ListContainers lists containers
func (c *Client) ListContainers(ctx context.Context, all bool) ([]ContainerInfo, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, err
	}

	var result []ContainerInfo
	for _, cont := range containers {
		info := ContainerInfo{
			ID:      cont.ID,
			Image:   cont.Image,
			Command: cont.Command,
			Created: formatTime(time.Unix(cont.Created, 0)),
			Status:  cont.Status,
			State:   cont.State,
		}

		if len(cont.Names) > 0 {
			info.Name = strings.TrimPrefix(cont.Names[0], "/")
		}

		// Health status
		if cont.Status != "" && strings.Contains(cont.Status, "(") {
			if strings.Contains(cont.Status, "healthy") {
				info.Health = "healthy"
			} else if strings.Contains(cont.Status, "unhealthy") {
				info.Health = "unhealthy"
			} else if strings.Contains(cont.Status, "starting") {
				info.Health = "starting"
			}
		}

		// Ports
		for _, port := range cont.Ports {
			info.Ports = append(info.Ports, PortMapping{
				IP:          port.IP,
				PrivatePort: port.PrivatePort,
				PublicPort:  port.PublicPort,
				Type:        port.Type,
			})
		}

		result = append(result, info)
	}

	return result, nil
}

// ImageInfo contains image information
type ImageInfo struct {
	ID         string
	Repository string
	Tag        string
	Digest     string
	Created    string
	CreatedAt  time.Time
	Size       int64
	Dangling   bool
}

// ListImages lists Docker images
func (c *Client) ListImages(ctx context.Context, all, danglingOnly bool) ([]ImageInfo, error) {
	opts := image.ListOptions{All: all}

	if danglingOnly {
		opts.Filters = filters.NewArgs()
		opts.Filters.Add("dangling", "true")
	}

	images, err := c.cli.ImageList(ctx, opts)
	if err != nil {
		return nil, err
	}

	var result []ImageInfo
	for _, img := range images {
		info := ImageInfo{
			ID:        strings.TrimPrefix(img.ID, "sha256:"),
			Size:      img.Size,
			CreatedAt: time.Unix(img.Created, 0),
			Created:   formatTime(time.Unix(img.Created, 0)),
			Dangling:  len(img.RepoTags) == 0,
		}

		if len(img.RepoTags) > 0 {
			parts := strings.Split(img.RepoTags[0], ":")
			info.Repository = parts[0]
			if len(parts) > 1 {
				info.Tag = parts[1]
			}
		} else {
			info.Repository = "<none>"
			info.Tag = "<none>"
		}

		if len(img.RepoDigests) > 0 {
			info.Digest = img.RepoDigests[0]
		}

		result = append(result, info)
	}

	return result, nil
}

// ContainerStats contains container statistics
type ContainerStats struct {
	ID            string
	Name          string
	CPUPercent    float64
	MemoryUsage   int64
	MemoryLimit   int64
	MemoryPercent float64
	NetInput      int64
	NetOutput     int64
	BlockInput    int64
	BlockOutput   int64
	PIDs          uint64
}

// GetContainerStats gets statistics for containers
func (c *Client) GetContainerStats(ctx context.Context, containers []ContainerInfo) ([]ContainerStats, error) {
	var result []ContainerStats

	for _, cont := range containers {
		stats, err := c.cli.ContainerStats(ctx, cont.ID, false)
		if err != nil {
			continue
		}

		var statsJSON types.StatsJSON
		decoder := json.NewDecoder(stats.Body)
		if err := decoder.Decode(&statsJSON); err != nil {
			stats.Body.Close()
			continue
		}
		stats.Body.Close()

		cs := ContainerStats{
			ID:   cont.ID,
			Name: cont.Name,
			PIDs: statsJSON.PidsStats.Current,
		}

		// Calculate CPU percent
		cpuDelta := float64(statsJSON.CPUStats.CPUUsage.TotalUsage - statsJSON.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(statsJSON.CPUStats.SystemUsage - statsJSON.PreCPUStats.SystemUsage)
		if systemDelta > 0 && cpuDelta > 0 {
			cs.CPUPercent = (cpuDelta / systemDelta) * float64(statsJSON.CPUStats.OnlineCPUs) * 100.0
		}

		// Memory
		cs.MemoryUsage = int64(statsJSON.MemoryStats.Usage)
		cs.MemoryLimit = int64(statsJSON.MemoryStats.Limit)
		if cs.MemoryLimit > 0 {
			cs.MemoryPercent = float64(cs.MemoryUsage) / float64(cs.MemoryLimit) * 100.0
		}

		// Network I/O
		for _, netStats := range statsJSON.Networks {
			cs.NetInput += int64(netStats.RxBytes)
			cs.NetOutput += int64(netStats.TxBytes)
		}

		// Block I/O
		for _, bioEntry := range statsJSON.BlkioStats.IoServiceBytesRecursive {
			switch bioEntry.Op {
			case "Read", "read":
				cs.BlockInput += int64(bioEntry.Value)
			case "Write", "write":
				cs.BlockOutput += int64(bioEntry.Value)
			}
		}

		result = append(result, cs)
	}

	return result, nil
}

// MountInfo contains mount information
type MountInfo struct {
	Type        string
	Name        string
	Source      string
	Destination string
	Driver      string
	Mode        string
	RW          bool
}

// NetworkInfo contains network information
type NetworkInfo struct {
	NetworkID  string
	IPAddress  string
	Gateway    string
	MacAddress string
}

// ContainerDetails contains detailed container information
type ContainerDetails struct {
	ID           string
	Name         string
	Image        string
	Created      string
	StartedAt    string
	FinishedAt   string
	State        string
	Status       string
	Health       string
	HealthLog    string
	RestartCount int
	Platform     string
	Command      string
	Entrypoint   string
	Env          []string
	Ports        []PortMapping
	Mounts       []MountInfo
	Networks     map[string]NetworkInfo
	Labels       map[string]string
}

// InspectContainer inspects a container
func (c *Client) InspectContainer(ctx context.Context, containerID string) (*ContainerDetails, error) {
	inspect, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}

	details := &ContainerDetails{
		ID:           inspect.ID,
		Name:         strings.TrimPrefix(inspect.Name, "/"),
		Image:        inspect.Config.Image,
		Created:      inspect.Created,
		StartedAt:    inspect.State.StartedAt,
		FinishedAt:   inspect.State.FinishedAt,
		State:        inspect.State.Status,
		Status:       inspect.State.Status,
		RestartCount: inspect.RestartCount,
		Platform:     inspect.Platform,
		Env:          inspect.Config.Env,
		Labels:       inspect.Config.Labels,
		Networks:     make(map[string]NetworkInfo),
	}

	// Command
	if len(inspect.Config.Cmd) > 0 {
		details.Command = strings.Join(inspect.Config.Cmd, " ")
	}
	if len(inspect.Config.Entrypoint) > 0 {
		details.Entrypoint = strings.Join(inspect.Config.Entrypoint, " ")
	}

	// Health
	if inspect.State.Health != nil {
		details.Health = inspect.State.Health.Status
		if len(inspect.State.Health.Log) > 0 {
			lastLog := inspect.State.Health.Log[len(inspect.State.Health.Log)-1]
			details.HealthLog = lastLog.Output
		}
	}

	// Ports
	for port, bindings := range inspect.NetworkSettings.Ports {
		pm := PortMapping{
			PrivatePort: uint16(port.Int()),
			Type:        port.Proto(),
		}
		if len(bindings) > 0 {
			pm.IP = bindings[0].HostIP
			fmt.Sscanf(bindings[0].HostPort, "%d", &pm.PublicPort)
		}
		details.Ports = append(details.Ports, pm)
	}

	// Mounts
	for _, mount := range inspect.Mounts {
		details.Mounts = append(details.Mounts, MountInfo{
			Type:        string(mount.Type),
			Name:        mount.Name,
			Source:      mount.Source,
			Destination: mount.Destination,
			Driver:      mount.Driver,
			Mode:        mount.Mode,
			RW:          mount.RW,
		})
	}

	// Networks
	for name, net := range inspect.NetworkSettings.Networks {
		details.Networks[name] = NetworkInfo{
			NetworkID:  net.NetworkID,
			IPAddress:  net.IPAddress,
			Gateway:    net.Gateway,
			MacAddress: net.MacAddress,
		}
	}

	return details, nil
}

// LogOptions contains log options
type LogOptions struct {
	Tail       int
	Follow     bool
	Timestamps bool
	Since      string
	Until      string
	Level      string
}

// LogLine represents a log line
type LogLine struct {
	Timestamp string
	Stream    string
	Content   string
	Level     string
}

// StreamLogs streams container logs
func (c *Client) StreamLogs(ctx context.Context, containerID string, opts LogOptions, callback func(LogLine)) error {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: opts.Timestamps,
		Follow:     opts.Follow,
		Tail:       fmt.Sprintf("%d", opts.Tail),
	}

	if opts.Since != "" {
		options.Since = opts.Since
	}
	if opts.Until != "" {
		options.Until = opts.Until
	}

	logs, err := c.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return err
	}
	defer logs.Close()

	reader := bufio.NewReader(logs)
	for {
		// Docker multiplexed stream format: [8]byte header + content
		header := make([]byte, 8)
		_, err := io.ReadFull(reader, header)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Get stream type and size
		streamType := header[0]
		size := int(header[4])<<24 | int(header[5])<<16 | int(header[6])<<8 | int(header[7])

		content := make([]byte, size)
		_, err = io.ReadFull(reader, content)
		if err != nil {
			return err
		}

		line := LogLine{
			Content: strings.TrimSpace(string(content)),
		}

		// Set stream type
		if streamType == 1 {
			line.Stream = "stdout"
		} else {
			line.Stream = "stderr"
		}

		// Parse timestamp if present
		if opts.Timestamps && len(line.Content) > 30 {
			parts := strings.SplitN(line.Content, " ", 2)
			if len(parts) == 2 {
				line.Timestamp = parts[0]
				line.Content = parts[1]
			}
		}

		// Detect log level
		line.Level = detectLogLevel(line.Content)

		// Filter by level if specified
		if opts.Level != "" && !matchesLevel(line.Level, opts.Level) {
			continue
		}

		callback(line)
	}

	return nil
}

func detectLogLevel(content string) string {
	lower := strings.ToLower(content)

	patterns := map[string]*regexp.Regexp{
		"error": regexp.MustCompile(`\b(error|err|fatal|panic|exception)\b`),
		"warn":  regexp.MustCompile(`\b(warn|warning)\b`),
		"info":  regexp.MustCompile(`\b(info)\b`),
		"debug": regexp.MustCompile(`\b(debug|trace)\b`),
	}

	for level, pattern := range patterns {
		if pattern.MatchString(lower) {
			return level
		}
	}

	return ""
}

func matchesLevel(detected, filter string) bool {
	filter = strings.ToLower(filter)
	detected = strings.ToLower(detected)

	if filter == detected {
		return true
	}

	// Include higher severity levels
	levels := []string{"debug", "info", "warn", "error"}
	filterIdx := -1
	detectedIdx := -1

	for i, l := range levels {
		if l == filter {
			filterIdx = i
		}
		if l == detected {
			detectedIdx = i
		}
	}

	return detectedIdx >= filterIdx
}

// NetworkDetails contains network details
type NetworkDetails struct {
	ID   string
	Name string
}

// VolumeDetails contains volume details
type VolumeDetails struct {
	Name string
	Size int64
}

// FindStoppedContainers finds stopped containers
func (c *Client) FindStoppedContainers(ctx context.Context) ([]ContainerInfo, error) {
	containers, err := c.ListContainers(ctx, true)
	if err != nil {
		return nil, err
	}

	var result []ContainerInfo
	for _, cont := range containers {
		if cont.State == "exited" || cont.State == "dead" {
			result = append(result, cont)
		}
	}
	return result, nil
}

// RemoveContainers removes containers
func (c *Client) RemoveContainers(ctx context.Context, containers []ContainerInfo) (int, int64, error) {
	deleted := 0
	for _, cont := range containers {
		err := c.cli.ContainerRemove(ctx, cont.ID, container.RemoveOptions{})
		if err == nil {
			deleted++
		}
	}
	return deleted, 0, nil
}

// FindUnusedImages finds unused images
func (c *Client) FindUnusedImages(ctx context.Context, all bool) ([]ImageInfo, error) {
	return c.ListImages(ctx, false, !all)
}

// RemoveImages removes images
func (c *Client) RemoveImages(ctx context.Context, images []ImageInfo) (int, int64, error) {
	deleted := 0
	var spaceReclaimed int64

	for _, img := range images {
		_, err := c.cli.ImageRemove(ctx, img.ID, image.RemoveOptions{})
		if err == nil {
			deleted++
			spaceReclaimed += img.Size
		}
	}
	return deleted, spaceReclaimed, nil
}

// FindUnusedNetworks finds unused networks
func (c *Client) FindUnusedNetworks(ctx context.Context) ([]NetworkDetails, error) {
	networks, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []NetworkDetails
	for _, net := range networks {
		// Skip default networks
		if net.Name == "bridge" || net.Name == "host" || net.Name == "none" {
			continue
		}

		// Check if network has no containers
		inspect, err := c.cli.NetworkInspect(ctx, net.ID, network.InspectOptions{})
		if err != nil {
			continue
		}

		if len(inspect.Containers) == 0 {
			result = append(result, NetworkDetails{
				ID:   net.ID,
				Name: net.Name,
			})
		}
	}

	return result, nil
}

// RemoveNetworks removes networks
func (c *Client) RemoveNetworks(ctx context.Context, networks []NetworkDetails) (int, error) {
	deleted := 0
	for _, net := range networks {
		err := c.cli.NetworkRemove(ctx, net.ID)
		if err == nil {
			deleted++
		}
	}
	return deleted, nil
}

// FindUnusedVolumes finds unused volumes
func (c *Client) FindUnusedVolumes(ctx context.Context) ([]VolumeDetails, error) {
	volumes, err := c.cli.VolumeList(ctx, volume.ListOptions{
		Filters: filters.NewArgs(filters.Arg("dangling", "true")),
	})
	if err != nil {
		return nil, err
	}

	var result []VolumeDetails
	for _, vol := range volumes.Volumes {
		result = append(result, VolumeDetails{
			Name: vol.Name,
			Size: vol.UsageData.Size,
		})
	}

	return result, nil
}

// RemoveVolumes removes volumes
func (c *Client) RemoveVolumes(ctx context.Context, volumes []VolumeDetails) (int, int64, error) {
	deleted := 0
	var spaceReclaimed int64

	for _, vol := range volumes {
		err := c.cli.VolumeRemove(ctx, vol.Name, false)
		if err == nil {
			deleted++
			spaceReclaimed += vol.Size
		}
	}
	return deleted, spaceReclaimed, nil
}

// GetBuildCacheSize gets build cache size
func (c *Client) GetBuildCacheSize(ctx context.Context) (int64, error) {
	usage, err := c.cli.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		return 0, err
	}

	var total int64
	if usage.BuildCache != nil {
		for _, bc := range usage.BuildCache {
			total += bc.Size
		}
	}

	return total, nil
}

// PruneBuildCache prunes build cache
func (c *Client) PruneBuildCache(ctx context.Context) (int64, error) {
	report, err := c.cli.BuildCachePrune(ctx, types.BuildCachePruneOptions{All: true})
	if err != nil {
		return 0, err
	}
	return int64(report.SpaceReclaimed), nil
}

func formatTime(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return fmt.Sprintf("%d seconds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	default:
		return fmt.Sprintf("%d weeks ago", int(d.Hours()/(24*7)))
	}
}


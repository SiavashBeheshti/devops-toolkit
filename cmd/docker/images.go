package docker

import (
	"context"
	"fmt"
	"sort"

	"github.com/beheshti/devops-toolkit/pkg/completion"
	"github.com/beheshti/devops-toolkit/pkg/docker"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "List and analyze images",
		Long: `List Docker images with enhanced analysis.

Features:
  • Size breakdown and visualization
  • Dangling image detection
  • Tag analysis
  • Layer count display`,
		RunE: runImages,
	}

	cmd.Flags().BoolP("all", "a", false, "Show all images (including intermediate)")
	cmd.Flags().Bool("dangling", false, "Show only dangling images")
	cmd.Flags().StringP("sort", "s", "size", "Sort by: name, size, created")
	cmd.Flags().Bool("digest", false, "Show image digests")

	// Register flag completions
	_ = cmd.RegisterFlagCompletionFunc("sort", completion.ImageSortCompletion)

	return cmd
}

func runImages(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Fetching images...")

	client, err := docker.NewClient()
	if err != nil {
		output.SpinnerError("Failed to connect to Docker")
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	showAll, _ := cmd.Flags().GetBool("all")
	danglingOnly, _ := cmd.Flags().GetBool("dangling")
	sortBy, _ := cmd.Flags().GetString("sort")
	showDigest, _ := cmd.Flags().GetBool("digest")

	images, err := client.ListImages(ctx, showAll, danglingOnly)
	if err != nil {
		output.SpinnerError("Failed to list images")
		return fmt.Errorf("failed to list images: %w", err)
	}

	output.SpinnerSuccess(fmt.Sprintf("Found %d images", len(images)))
	output.Newline()

	if len(images) == 0 {
		output.Info("No images found")
		return nil
	}

	// Sort images
	sortImages(images, sortBy)

	// Calculate total size
	var totalSize int64
	var danglingCount int
	for _, img := range images {
		totalSize += img.Size
		if img.Dangling {
			danglingCount++
		}
	}

	// Build table
	headers := []string{"Repository", "Tag", "Image ID", "Created", "Size"}
	if showDigest {
		headers = append(headers, "Digest")
	}

	table := output.NewTable(output.TableConfig{
		Title:      "Docker Images",
		Headers:    headers,
		ShowBorder: true,
	})

	for _, img := range images {
		repo := img.Repository
		tag := img.Tag

		if img.Dangling {
			repo = "<none>"
			tag = "<none>"
		}

		row := []string{
			repo,
			tag,
			truncateID(img.ID),
			img.Created,
			formatSize(img.Size),
		}

		if showDigest {
			digest := img.Digest
			if len(digest) > 20 {
				digest = digest[:17] + "..."
			}
			row = append(row, digest)
		}

		colors := getImageRowColors(img, showDigest)
		table.AddColoredRow(row, colors)
	}

	table.Render()

	// Summary
	output.Newline()
	output.Print(output.Section("Image Summary"))
	output.Printf("  Total Images: %d\n", len(images))
	output.Printf("  Total Size: %s\n", formatSize(totalSize))
	if danglingCount > 0 {
		output.Printf("  %s Dangling: %d (reclaimable space)\n",
			output.WarningStyle.Render(output.IconWarning), danglingCount)
	}

	// Size breakdown by repository
	output.Newline()
	output.Print(output.SubSection("Size by Repository"))
	repoSizes := make(map[string]int64)
	for _, img := range images {
		if !img.Dangling {
			repoSizes[img.Repository] += img.Size
		}
	}

	// Sort by size
	type repoSize struct {
		name string
		size int64
	}
	var sorted []repoSize
	for name, size := range repoSizes {
		sorted = append(sorted, repoSize{name, size})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].size > sorted[j].size
	})

	// Show top 5
	for i, rs := range sorted {
		if i >= 5 {
			break
		}
		bar := output.ProgressBar(int(float64(rs.size)/float64(totalSize)*100), 100, 20)
		output.Printf("    %s: %s %s\n", rs.name, formatSize(rs.size), bar)
	}

	output.Newline()
	return nil
}

func sortImages(images []docker.ImageInfo, sortBy string) {
	sort.Slice(images, func(i, j int) bool {
		switch sortBy {
		case "name":
			return images[i].Repository < images[j].Repository
		case "created":
			return images[i].CreatedAt.After(images[j].CreatedAt)
		default: // size
			return images[i].Size > images[j].Size
		}
	})
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func getImageRowColors(img docker.ImageInfo, showDigest bool) []tablewriter.Colors {
	var repoColor, tagColor int

	if img.Dangling {
		repoColor = tablewriter.FgYellowColor
		tagColor = tablewriter.FgYellowColor
	} else {
		repoColor = tablewriter.FgCyanColor
		tagColor = tablewriter.FgGreenColor
	}

	colors := []tablewriter.Colors{
		{repoColor},                      // Repository
		{tagColor},                       // Tag
		{tablewriter.FgHiBlackColor},     // ID
		{tablewriter.FgHiBlackColor},     // Created
		{tablewriter.FgYellowColor},      // Size
	}

	if showDigest {
		colors = append(colors, tablewriter.Colors{tablewriter.FgHiBlackColor})
	}

	return colors
}


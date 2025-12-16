package gitlabclient

import (
	"fmt"
	"time"

	"github.com/xanzy/go-gitlab"
)

// Client wraps the GitLab client
type Client struct {
	client *gitlab.Client
}

// NewClient creates a new GitLab client
func NewClient(url, token string) (*Client, error) {
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(url))
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %w", err)
	}

	return &Client{client: client}, nil
}

// PipelineInfo contains pipeline information
type PipelineInfo struct {
	ID        int
	Status    string
	Ref       string
	SHA       string
	WebURL    string
	CreatedAt string
	Duration  string
}

// PipelineFilter contains filter options
type PipelineFilter struct {
	Status string
	Ref    string
	Limit  int
}

// ListPipelines lists pipelines
func (c *Client) ListPipelines(projectID string, filter PipelineFilter) ([]PipelineInfo, error) {
	opts := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: filter.Limit,
		},
	}

	if filter.Status != "" {
		status := gitlab.BuildStateValue(filter.Status)
		opts.Status = &status
	}
	if filter.Ref != "" {
		opts.Ref = &filter.Ref
	}

	pipelines, _, err := c.client.Pipelines.ListProjectPipelines(projectID, opts)
	if err != nil {
		return nil, err
	}

	var result []PipelineInfo
	for _, pl := range pipelines {
		info := PipelineInfo{
			ID:     pl.ID,
			Status: pl.Status,
			Ref:    pl.Ref,
			SHA:    pl.SHA,
			WebURL: pl.WebURL,
		}

		if pl.CreatedAt != nil {
			info.CreatedAt = formatTime(*pl.CreatedAt)
		}

		// Get duration from detailed pipeline info
		detailed, _, err := c.client.Pipelines.GetPipeline(projectID, pl.ID)
		if err == nil && detailed.Duration > 0 {
			info.Duration = formatDuration(float64(detailed.Duration))
		}

		result = append(result, info)
	}

	return result, nil
}

// JobInfo contains job information
type JobInfo struct {
	ID        int
	Name      string
	Stage     string
	Status    string
	Duration  string
	StartedAt string
	WebURL    string
}

// JobFilter contains job filter options
type JobFilter struct {
	Status string
	Stage  string
}

// ListPipelineJobs lists pipeline jobs
func (c *Client) ListPipelineJobs(projectID string, pipelineID int, filter JobFilter) ([]JobInfo, error) {
	opts := &gitlab.ListJobsOptions{}

	jobs, _, err := c.client.Jobs.ListPipelineJobs(projectID, pipelineID, opts)
	if err != nil {
		return nil, err
	}

	var result []JobInfo
	for _, job := range jobs {
		// Apply filters
		if filter.Status != "" && job.Status != filter.Status {
			continue
		}
		if filter.Stage != "" && job.Stage != filter.Stage {
			continue
		}

		info := JobInfo{
			ID:     job.ID,
			Name:   job.Name,
			Stage:  job.Stage,
			Status: job.Status,
			WebURL: job.WebURL,
		}

		if job.Duration > 0 {
			info.Duration = formatDuration(float64(job.Duration))
		}

		if job.StartedAt != nil {
			info.StartedAt = formatTime(*job.StartedAt)
		}

		result = append(result, info)
	}

	return result, nil
}

// TriggerPipeline triggers a new pipeline
func (c *Client) TriggerPipeline(projectID, ref string, variables map[string]string) (*PipelineInfo, error) {
	opts := &gitlab.CreatePipelineOptions{
		Ref: &ref,
	}

	// Add variables
	if len(variables) > 0 {
		vars := make([]*gitlab.PipelineVariableOptions, 0, len(variables))
		for k, v := range variables {
			key := k
			value := v
			vars = append(vars, &gitlab.PipelineVariableOptions{
				Key:   &key,
				Value: &value,
			})
		}
		opts.Variables = &vars
	}

	pipeline, _, err := c.client.Pipelines.CreatePipeline(projectID, opts)
	if err != nil {
		return nil, err
	}

	return &PipelineInfo{
		ID:     pipeline.ID,
		Status: pipeline.Status,
		Ref:    pipeline.Ref,
		SHA:    pipeline.SHA,
		WebURL: pipeline.WebURL,
	}, nil
}

// WaitForPipeline waits for pipeline to complete
func (c *Client) WaitForPipeline(projectID string, pipelineID int) (*PipelineInfo, error) {
	for {
		pipeline, _, err := c.client.Pipelines.GetPipeline(projectID, pipelineID)
		if err != nil {
			return nil, err
		}

		// Check if pipeline is finished
		switch pipeline.Status {
		case "success", "failed", "canceled", "skipped":
			return &PipelineInfo{
				ID:       pipeline.ID,
				Status:   pipeline.Status,
				Ref:      pipeline.Ref,
				SHA:      pipeline.SHA,
				WebURL:   pipeline.WebURL,
				Duration: formatDuration(float64(pipeline.Duration)),
			}, nil
		}

		time.Sleep(5 * time.Second)
	}
}

// ArtifactInfo contains artifact information
type ArtifactInfo struct {
	JobID    int
	JobName  string
	Filename string
	Size     int64
	ExpireAt string
}

// GetJobArtifacts gets artifacts for a job
func (c *Client) GetJobArtifacts(projectID string, jobID int) (*ArtifactInfo, error) {
	job, _, err := c.client.Jobs.GetJob(projectID, jobID)
	if err != nil {
		return nil, err
	}

	if len(job.Artifacts) == 0 {
		return nil, nil
	}

	info := &ArtifactInfo{
		JobID:   job.ID,
		JobName: job.Name,
	}

	for _, art := range job.Artifacts {
		info.Filename = art.Filename
		info.Size = int64(art.Size)
	}

	if job.ArtifactsExpireAt != nil {
		info.ExpireAt = formatTime(*job.ArtifactsExpireAt)
	}

	return info, nil
}

// ListPipelineArtifacts lists all artifacts from a pipeline
func (c *Client) ListPipelineArtifacts(projectID string, pipelineID int) ([]ArtifactInfo, error) {
	jobs, _, err := c.client.Jobs.ListPipelineJobs(projectID, pipelineID, nil)
	if err != nil {
		return nil, err
	}

	var result []ArtifactInfo
	for _, job := range jobs {
		if len(job.Artifacts) > 0 {
			for _, art := range job.Artifacts {
				info := ArtifactInfo{
					JobID:    job.ID,
					JobName:  job.Name,
					Filename: art.Filename,
					Size:     int64(art.Size),
				}
				if job.ArtifactsExpireAt != nil {
					info.ExpireAt = formatTime(*job.ArtifactsExpireAt)
				}
				result = append(result, info)
			}
		}
	}

	return result, nil
}

// ProjectInfo contains project information
type ProjectInfo struct {
	ID                int
	Name              string
	PathWithNamespace string
	DefaultBranch     string
	WebURL            string
}

// GetProject gets project information
func (c *Client) GetProject(projectID string) (*ProjectInfo, error) {
	project, _, err := c.client.Projects.GetProject(projectID, nil)
	if err != nil {
		return nil, err
	}

	return &ProjectInfo{
		ID:                project.ID,
		Name:              project.Name,
		PathWithNamespace: project.PathWithNamespace,
		DefaultBranch:     project.DefaultBranch,
		WebURL:            project.WebURL,
	}, nil
}

// GetLatestPipeline gets the latest pipeline for a ref
func (c *Client) GetLatestPipeline(projectID, ref string) (*PipelineInfo, error) {
	opts := &gitlab.ListProjectPipelinesOptions{
		Ref: &ref,
		ListOptions: gitlab.ListOptions{
			PerPage: 1,
		},
	}

	pipelines, _, err := c.client.Pipelines.ListProjectPipelines(projectID, opts)
	if err != nil {
		return nil, err
	}

	if len(pipelines) == 0 {
		return nil, fmt.Errorf("no pipelines found")
	}

	pl := pipelines[0]
	detailed, _, err := c.client.Pipelines.GetPipeline(projectID, pl.ID)
	if err != nil {
		return nil, err
	}

	return &PipelineInfo{
		ID:       detailed.ID,
		Status:   detailed.Status,
		Ref:      detailed.Ref,
		SHA:      detailed.SHA,
		WebURL:   detailed.WebURL,
		Duration: formatDuration(float64(detailed.Duration)),
	}, nil
}

// PipelineStats contains pipeline statistics
type PipelineStats struct {
	Success     int
	Failed      int
	Other       int
	AvgDuration string
}

// GetPipelineStats gets pipeline statistics
func (c *Client) GetPipelineStats(projectID string) (*PipelineStats, error) {
	// Get pipelines from last 30 days
	since := time.Now().AddDate(0, 0, -30)
	opts := &gitlab.ListProjectPipelinesOptions{
		UpdatedAfter: &since,
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	pipelines, _, err := c.client.Pipelines.ListProjectPipelines(projectID, opts)
	if err != nil {
		return nil, err
	}

	stats := &PipelineStats{}
	var totalDuration float64
	var durationCount int

	for _, pl := range pipelines {
		switch pl.Status {
		case "success":
			stats.Success++
		case "failed":
			stats.Failed++
		default:
			stats.Other++
		}

		// Get duration
		detailed, _, err := c.client.Pipelines.GetPipeline(projectID, pl.ID)
		if err == nil && detailed.Duration > 0 {
			totalDuration += float64(detailed.Duration)
			durationCount++
		}
	}

	if durationCount > 0 {
		avgDuration := totalDuration / float64(durationCount)
		stats.AvgDuration = formatDuration(avgDuration)
	}

	return stats, nil
}

// EnvironmentInfo contains environment information
type EnvironmentInfo struct {
	ID             int
	Name           string
	State          string
	ExternalURL    string
	LastDeployment string
}

// ListEnvironments lists project environments
func (c *Client) ListEnvironments(projectID string) ([]EnvironmentInfo, error) {
	envs, _, err := c.client.Environments.ListEnvironments(projectID, nil)
	if err != nil {
		return nil, err
	}

	var result []EnvironmentInfo
	for _, env := range envs {
		info := EnvironmentInfo{
			ID:          env.ID,
			Name:        env.Name,
			State:       env.State,
			ExternalURL: env.ExternalURL,
		}

		// Get last deployment
		if env.LastDeployment != nil {
			if env.LastDeployment.CreatedAt != nil {
				info.LastDeployment = formatTime(*env.LastDeployment.CreatedAt)
			}
		}

		result = append(result, info)
	}

	return result, nil
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
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}

func formatDuration(seconds float64) string {
	d := time.Duration(seconds) * time.Second

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}


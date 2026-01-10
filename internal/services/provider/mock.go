package provider

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type MockProvider struct {
	name    string
	jobs    map[string]*mockJob
	storage Storage
}

type mockJob struct {
	status      JobStatus
	progress    int
	resultURL   *string
	error       *string
	createdAt   time.Time
	completedAt *time.Time
}

func NewMockProvider(name string, storage Storage) *MockProvider {
	return &MockProvider{
		name:    name,
		jobs:    make(map[string]*mockJob),
		storage: storage,
	}
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) Submit(ctx context.Context, req UnifiedGenRequest) (*SubmitResult, error) {
	jobID := fmt.Sprintf("mock_job_%d_%s", time.Now().UnixNano(), req.RequestID)

	p.jobs[jobID] = &mockJob{
		status:    JobStatusPending,
		progress:  0,
		createdAt: time.Now(),
	}

	go p.simulateJob(jobID, req)

	return &SubmitResult{
		ProviderJobID: jobID,
	}, nil
}

func (p *MockProvider) Status(ctx context.Context, jobID string) (*StatusResult, error) {
	job, exists := p.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return &StatusResult{
		Status:    job.status,
		Progress:  job.progress,
		ResultURL: job.resultURL,
		Error:     job.error,
	}, nil
}

func (p *MockProvider) simulateJob(jobID string, req UnifiedGenRequest) {
	job := p.jobs[jobID]

	time.Sleep(1 * time.Second)

	job.status = JobStatusRunning
	job.progress = 20

	steps := []int{40, 60, 80, 100}
	for _, step := range steps {
		time.Sleep(time.Duration(rand.Intn(5)+3) * time.Second)
		job.progress = step
	}

	resultURL, err := p.generateMockResult(req)
	if err != nil {
		job.status = JobStatusFailed
		errMsg := fmt.Sprintf("failed to generate result: %v", err)
		job.error = &errMsg
		now := time.Now()
		job.completedAt = &now
		return
	}

	job.status = JobStatusSucceeded
	job.progress = 100
	job.resultURL = &resultURL
	now := time.Now()
	job.completedAt = &now
}

func (p *MockProvider) generateMockResult(req UnifiedGenRequest) (string, error) {
	objectKey := fmt.Sprintf("mock/%s/%d", req.Type, time.Now().UnixNano())

	content := fmt.Sprintf("Mock generated %s for request: %s\nPrompt: %s",
		req.Type, req.RequestID, req.Prompt)

	url, err := p.storage.Upload(context.Background(), objectKey, []byte(content), "text/plain")
	if err != nil {
		return "", err
	}

	return url, nil
}

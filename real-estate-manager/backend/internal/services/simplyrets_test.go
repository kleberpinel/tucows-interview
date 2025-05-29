package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"real-estate-manager/backend/internal/mocks"
	"real-estate-manager/backend/internal/models"

	"go.uber.org/mock/gomock"
)

func TestNewSimplyRETSService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPropertyRepository(ctrl)
	service := NewSimplyRETSService(mockRepo)

	if service == nil {
		t.Error("NewSimplyRETSService() returned nil")
	}
	if service.propertyRepo != mockRepo {
		t.Error("NewSimplyRETSService() did not set repository correctly")
	}
	if service.client == nil {
		t.Error("NewSimplyRETSService() did not set HTTP client")
	}
	if service.baseURL != "https://api.simplyrets.com" {
		t.Errorf("Expected baseURL to be 'https://api.simplyrets.com', got '%s'", service.baseURL)
	}
	if service.username != "simplyrets" {
		t.Errorf("Expected username to be 'simplyrets', got '%s'", service.username)
	}
	if service.password != "simplyrets" {
		t.Errorf("Expected password to be 'simplyrets', got '%s'", service.password)
	}
	if service.imagesDir != "./uploads/images" {
		t.Errorf("Expected imagesDir to be './uploads/images', got '%s'", service.imagesDir)
	}
}

func TestJobManager_AddJob(t *testing.T) {
	tests := []struct {
		name   string
		jobID  string
		job    *ProcessingJob
		verify func(t *testing.T, jm *JobManager, jobID string, job *ProcessingJob)
	}{
		{
			name:  "successful job addition",
			jobID: "test-job-1",
			job: &ProcessingJob{
				ID:        "test-job-1",
				Status:    make(chan models.ProcessingStatus, 10),
				StartTime: time.Now(),
			},
			verify: func(t *testing.T, jm *JobManager, jobID string, job *ProcessingJob) {
				retrievedJob, exists := jm.GetJob(jobID)
				if !exists {
					t.Error("Job was not added to manager")
				}
				if retrievedJob.ID != job.ID {
					t.Errorf("Expected job ID %s, got %s", job.ID, retrievedJob.ID)
				}
			},
		},
		{
			name:  "add multiple jobs",
			jobID: "test-job-2",
			job: &ProcessingJob{
				ID:        "test-job-2",
				Status:    make(chan models.ProcessingStatus, 10),
				StartTime: time.Now(),
			},
			verify: func(t *testing.T, jm *JobManager, jobID string, job *ProcessingJob) {
				if len(jm.jobs) < 1 {
					t.Error("Job count should be at least 1")
				}
			},
		},
	}

	jm := NewJobManager()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm.AddJob(tt.jobID, tt.job)
			tt.verify(t, jm, tt.jobID, tt.job)
		})
	}
}

func TestJobManager_GetJob(t *testing.T) {
	tests := []struct {
		name       string
		jobID      string
		setupJobs  func(jm *JobManager)
		expectJob  bool
		verifyJob  func(t *testing.T, job *ProcessingJob)
	}{
		{
			name:  "get existing job",
			jobID: "existing-job",
			setupJobs: func(jm *JobManager) {
				job := &ProcessingJob{
					ID:        "existing-job",
					Status:    make(chan models.ProcessingStatus, 10),
					StartTime: time.Now(),
				}
				jm.AddJob("existing-job", job)
			},
			expectJob: true,
			verifyJob: func(t *testing.T, job *ProcessingJob) {
				if job.ID != "existing-job" {
					t.Errorf("Expected job ID 'existing-job', got '%s'", job.ID)
				}
			},
		},
		{
			name:      "get non-existent job",
			jobID:     "non-existent",
			setupJobs: func(jm *JobManager) {},
			expectJob: false,
			verifyJob: func(t *testing.T, job *ProcessingJob) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := NewJobManager()
			tt.setupJobs(jm)

			job, exists := jm.GetJob(tt.jobID)
			if exists != tt.expectJob {
				t.Errorf("Expected job existence %t, got %t", tt.expectJob, exists)
			}

			if tt.expectJob {
				tt.verifyJob(t, job)
			}
		})
	}
}

func TestJobManager_RemoveJob(t *testing.T) {
	tests := []struct {
		name      string
		jobID     string
		setupJobs func(jm *JobManager)
		verify    func(t *testing.T, jm *JobManager, jobID string)
	}{
		{
			name:  "remove existing job",
			jobID: "job-to-remove",
			setupJobs: func(jm *JobManager) {
				job := &ProcessingJob{
					ID:        "job-to-remove",
					Status:    make(chan models.ProcessingStatus, 10),
					StartTime: time.Now(),
				}
				jm.AddJob("job-to-remove", job)
			},
			verify: func(t *testing.T, jm *JobManager, jobID string) {
				_, exists := jm.GetJob(jobID)
				if exists {
					t.Error("Job should have been removed")
				}
			},
		},
		{
			name:      "remove non-existent job",
			jobID:     "non-existent",
			setupJobs: func(jm *JobManager) {},
			verify: func(t *testing.T, jm *JobManager, jobID string) {
				// Should not panic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := NewJobManager()
			tt.setupJobs(jm)

			jm.RemoveJob(tt.jobID)
			tt.verify(t, jm, tt.jobID)
		})
	}
}

func TestJobManager_MarkJobCompleted(t *testing.T) {
	tests := []struct {
		name        string
		jobID       string
		finalStatus models.ProcessingStatus
		setupJobs   func(jm *JobManager)
		verify      func(t *testing.T, jm *JobManager, jobID string, status models.ProcessingStatus)
	}{
		{
			name:  "mark existing job as completed",
			jobID: "job-to-complete",
			finalStatus: models.ProcessingStatus{
				Status:          "completed",
				TotalProperties: 10,
				ProcessedCount:  10,
				FailedCount:     0,
			},
			setupJobs: func(jm *JobManager) {
				job := &ProcessingJob{
					ID:        "job-to-complete",
					Status:    make(chan models.ProcessingStatus, 10),
					StartTime: time.Now(),
				}
				jm.AddJob("job-to-complete", job)
			},
			verify: func(t *testing.T, jm *JobManager, jobID string, status models.ProcessingStatus) {
				job, exists := jm.GetJob(jobID)
				if !exists {
					t.Error("Job should still exist after completion")
				}
				job.mu.RLock()
				if job.LastStatus == nil {
					t.Error("LastStatus should be set")
				} else if job.LastStatus.Status != status.Status {
					t.Errorf("Expected status '%s', got '%s'", status.Status, job.LastStatus.Status)
				}
				if job.CompletedAt == nil {
					t.Error("CompletedAt should be set")
				}
				job.mu.RUnlock()
			},
		},
		{
			name:  "mark non-existent job as completed",
			jobID: "non-existent",
			finalStatus: models.ProcessingStatus{
				Status: "completed",
			},
			setupJobs: func(jm *JobManager) {},
			verify: func(t *testing.T, jm *JobManager, jobID string, status models.ProcessingStatus) {
				// Should not panic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := NewJobManager()
			tt.setupJobs(jm)

			jm.MarkJobCompleted(tt.jobID, tt.finalStatus)
			tt.verify(t, jm, tt.jobID, tt.finalStatus)
		})
	}
}

func TestSimplyRETSService_StartPropertyProcessing(t *testing.T) {
	tests := []struct {
		name        string
		jobID       string
		limit       int
		setupMock   func(mock *mocks.MockPropertyRepository)
		expectError bool
		errorMsg    string
		verify      func(t *testing.T, jobID string)
	}{
		{
			name:  "successful processing start",
			jobID: "test-job-start",
			limit: 5,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// Mock will be called during actual processing in goroutine
			},
			expectError: false,
			verify: func(t *testing.T, jobID string) {
				// Verify job was added to manager
				job, exists := GlobalJobManager.GetJob(jobID)
				if !exists {
					t.Error("Job should exist in manager")
				}
				if job.ID != jobID {
					t.Errorf("Expected job ID %s, got %s", jobID, job.ID)
				}
				
				// Wait a bit for processing to start and then cancel to clean up
				time.Sleep(10 * time.Millisecond)
				if job.Cancel != nil {
					job.Cancel()
				}
				
				// Wait for job to be removed or timeout
				timeout := time.After(100 * time.Millisecond)
				ticker := time.NewTicker(5 * time.Millisecond)
				defer ticker.Stop()
				
				for {
					select {
					case <-timeout:
						// Force cleanup if timeout
						GlobalJobManager.RemoveJob(jobID)
						return
					case <-ticker.C:
						if _, exists := GlobalJobManager.GetJob(jobID); !exists {
							return
						}
					}
				}
			},
		},
		{
			name:  "processing start with zero limit",
			jobID: "test-job-zero",
			limit: 0,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// Mock will be called during actual processing
			},
			expectError: false,
			verify: func(t *testing.T, jobID string) {
				// Wait a bit for processing to start and then cancel to clean up
				time.Sleep(10 * time.Millisecond)
				job, exists := GlobalJobManager.GetJob(jobID)
				if exists && job.Cancel != nil {
					job.Cancel()
				}
				
				// Wait for job to be removed or timeout
				timeout := time.After(100 * time.Millisecond)
				ticker := time.NewTicker(5 * time.Millisecond)
				defer ticker.Stop()
				
				for {
					select {
					case <-timeout:
						// Force cleanup if timeout
						GlobalJobManager.RemoveJob(jobID)
						return
					case <-ticker.C:
						if _, exists := GlobalJobManager.GetJob(jobID); !exists {
							return
						}
					}
				}
			},
		},
	}

	// Create a mock HTTP server for API calls
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty array to prevent actual processing
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			tt.setupMock(mockRepo)

			service := NewSimplyRETSService(mockRepo)
			service.baseURL = server.URL // Use test server
			ctx := context.Background()

			err := service.StartPropertyProcessing(ctx, tt.jobID, tt.limit)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if !tt.expectError {
				tt.verify(t, tt.jobID)
			}
		})
	}
}

func TestSimplyRETSService_GetJobStatus(t *testing.T) {
	tests := []struct {
		name        string
		jobID       string
		setupJob    func() *ProcessingJob
		expectFound bool
		verifyStatus func(t *testing.T, status *models.ProcessingStatus)
	}{
		{
			name:  "get status for running job",
			jobID: "running-job",
			setupJob: func() *ProcessingJob {
				statusChan := make(chan models.ProcessingStatus, 10)
				job := &ProcessingJob{
					ID:        "running-job",
					Status:    statusChan,
					StartTime: time.Now(),
				}
				
				// Send a status update
				status := models.ProcessingStatus{
					Status:          "running",
					TotalProperties: 10,
					ProcessedCount:  5,
					StartedAt:       job.StartTime,
				}
				statusChan <- status
				
				return job
			},
			expectFound: true,
			verifyStatus: func(t *testing.T, status *models.ProcessingStatus) {
				if status.Status != "running" {
					t.Errorf("Expected status 'running', got '%s'", status.Status)
				}
				if status.ProcessedCount != 5 {
					t.Errorf("Expected processed count 5, got %d", status.ProcessedCount)
				}
			},
		},
		{
			name:  "get status for completed job",
			jobID: "completed-job",
			setupJob: func() *ProcessingJob {
				job := &ProcessingJob{
					ID:        "completed-job",
					Status:    make(chan models.ProcessingStatus, 10),
					StartTime: time.Now(),
				}
				
				// Set completed status
				completedStatus := models.ProcessingStatus{
					Status:         "completed",
					TotalProperties: 10,
					ProcessedCount: 10,
					StartedAt:      job.StartTime,
				}
				job.LastStatus = &completedStatus
				now := time.Now()
				job.CompletedAt = &now
				
				return job
			},
			expectFound: true,
			verifyStatus: func(t *testing.T, status *models.ProcessingStatus) {
				if status.Status != "completed" {
					t.Errorf("Expected status 'completed', got '%s'", status.Status)
				}
				if status.ProcessedCount != 10 {
					t.Errorf("Expected processed count 10, got %d", status.ProcessedCount)
				}
			},
		},
		{
			name:  "get status for non-existent job",
			jobID: "non-existent",
			setupJob: func() *ProcessingJob {
				return nil
			},
			expectFound: false,
			verifyStatus: func(t *testing.T, status *models.ProcessingStatus) {
				// Should not be called
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			service := NewSimplyRETSService(mockRepo)

			// Setup job if needed
			if job := tt.setupJob(); job != nil {
				GlobalJobManager.AddJob(tt.jobID, job)
				defer GlobalJobManager.RemoveJob(tt.jobID)
			}

			status, found := service.GetJobStatus(tt.jobID)

			if found != tt.expectFound {
				t.Errorf("Expected found %t, got %t", tt.expectFound, found)
			}

			if tt.expectFound {
				if status == nil {
					t.Error("Expected status but got nil")
				} else {
					tt.verifyStatus(t, status)
				}
			} else {
				if status != nil {
					t.Error("Expected nil status but got value")
				}
			}
		})
	}
}

func TestSimplyRETSService_CancelJob(t *testing.T) {
	tests := []struct {
		name      string
		jobID     string
		setupJob  func() *ProcessingJob
		expectSuccess bool
	}{
		{
			name:  "cancel existing job",
			jobID: "job-to-cancel",
			setupJob: func() *ProcessingJob {
				_, cancel := context.WithCancel(context.Background())
				return &ProcessingJob{
					ID:        "job-to-cancel",
					Status:    make(chan models.ProcessingStatus, 10),
					Cancel:    cancel,
					StartTime: time.Now(),
				}
			},
			expectSuccess: true,
		},
		{
			name:  "cancel non-existent job",
			jobID: "non-existent",
			setupJob: func() *ProcessingJob {
				return nil
			},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			service := NewSimplyRETSService(mockRepo)

			// Setup job if needed
			if job := tt.setupJob(); job != nil {
				GlobalJobManager.AddJob(tt.jobID, job)
			}

			success := service.CancelJob(tt.jobID)

			if success != tt.expectSuccess {
				t.Errorf("Expected success %t, got %t", tt.expectSuccess, success)
			}

			// Verify job was removed if cancelled successfully
			if tt.expectSuccess {
				_, exists := GlobalJobManager.GetJob(tt.jobID)
				if exists {
					t.Error("Job should have been removed after cancellation")
				}
			}
		})
	}
}

func TestSimplyRETSService_fetchProperties(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		serverResponse func() *httptest.Server
		expectError    bool
		errorMsg       string
		verifyResult   func(t *testing.T, properties []models.SimplyRETSProperty)
	}{
		{
			name:  "successful fetch with valid response",
			limit: 2,
			serverResponse: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify request
					if r.URL.Query().Get("limit") != "2" {
						t.Errorf("Expected limit=2, got %s", r.URL.Query().Get("limit"))
					}
					
					// Mock response
					properties := []models.SimplyRETSProperty{
						{
							ListingID: "123",
							MLSNumber: "MLS123",
							Address: models.SimplyRETSAddress{
								Full:         "123 Main St, City, State",
								StreetNumber: "123",
								StreetName:   "Main St",
							},
							ListPrice: 500000.0,
							Property: models.SimplyRETSPropertyDetails{
								PropertyType: "Single Family",
								Bedrooms:     3,
								Bathrooms:    2,
								Area:         2000,
							},
							Photos:  []string{"photo1.jpg", "photo2.jpg"},
							Remarks: "Beautiful house",
						},
					}
					
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(properties)
				}))
			},
			expectError: false,
			verifyResult: func(t *testing.T, properties []models.SimplyRETSProperty) {
				if len(properties) != 1 {
					t.Errorf("Expected 1 property, got %d", len(properties))
				}
				if properties[0].ListingID != "123" {
					t.Errorf("Expected ListingID '123', got '%s'", properties[0].ListingID)
				}
			},
		},
		{
			name:  "server returns 500 error",
			limit: 1,
			serverResponse: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectError: true,
			errorMsg:    "API returned status 500",
			verifyResult: func(t *testing.T, properties []models.SimplyRETSProperty) {
				// Should not be called
			},
		},
		{
			name:  "invalid JSON response",
			limit: 1,
			serverResponse: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte("invalid json"))
				}))
			},
			expectError: true,
			errorMsg:    "failed to decode response",
			verifyResult: func(t *testing.T, properties []models.SimplyRETSProperty) {
				// Should not be called
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			service := NewSimplyRETSService(mockRepo)

			// Setup test server
			server := tt.serverResponse()
			defer server.Close()

			// Override service baseURL to use test server
			service.baseURL = server.URL

			ctx := context.Background()
			properties, err := service.fetchProperties(ctx, tt.limit)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				tt.verifyResult(t, properties)
			}
		})
	}
}

func TestSimplyRETSService_processProperty(t *testing.T) {
	tests := []struct {
		name          string
		property      models.SimplyRETSProperty
		setupMock     func(mock *mocks.MockPropertyRepository)
		setupServer   func() *httptest.Server
		expectError   bool
		errorMsg      string
	}{
		{
			name: "successful property processing",
			property: models.SimplyRETSProperty{
				ListingID: "test-123",
				MLSNumber: "MLS123",
				Address: models.SimplyRETSAddress{
					Full:         "123 Test St, Test City, TS",
					StreetNumber: "123",
					StreetName:   "Test St",
				},
				ListPrice: 300000.0,
				Property: models.SimplyRETSPropertyDetails{
					PropertyType: "Condo",
					Bedrooms:     2,
					Bathrooms:    1,
					Area:         1200,
				},
				Photos:  []string{},
				Remarks: "Nice condo",
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			setupServer: func() *httptest.Server {
				// No images to download
				return nil
			},
			expectError: false,
		},
		{
			name: "property processing with repository error",
			property: models.SimplyRETSProperty{
				ListingID: "test-456",
				MLSNumber: "MLS456",
				Address: models.SimplyRETSAddress{
					Full:         "456 Test Ave, Test City, TS",
					StreetNumber: "456",
					StreetName:   "Test Ave",
				},
				ListPrice: 400000.0,
				Photos:    []string{},
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			setupServer: func() *httptest.Server {
				return nil
			},
			expectError: true,
			errorMsg:    "failed to save property test-456: database error",
		},
	}

	// Create temporary directory for images
	tempDir, err := os.MkdirTemp("", "simplyrets_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			tt.setupMock(mockRepo)

			service := NewSimplyRETSService(mockRepo)
			service.imagesDir = tempDir

			if tt.setupServer != nil {
				server := tt.setupServer()
				if server != nil {
					defer server.Close()
				}
			}

			ctx := context.Background()
			err := service.processProperty(ctx, tt.property)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestSimplyRETSService_downloadImages(t *testing.T) {
	tests := []struct {
		name         string
		imageURLs    []string
		propertyID   string
		setupServer  func() *httptest.Server
		expectError  bool
		errorMsg     string
		verifyResult func(t *testing.T, photos models.PhotoList)
	}{
		{
			name:       "successful image download",
			imageURLs:  []string{"/image1.jpg", "/image2.jpg"},
			propertyID: "prop123",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "image/jpeg")
					w.Write([]byte("fake image data"))
				}))
			},
			expectError: false,
			verifyResult: func(t *testing.T, photos models.PhotoList) {
				if len(photos) != 2 {
					t.Errorf("Expected 2 photos, got %d", len(photos))
				}
				
				// Check that we have the expected captions (order may vary due to concurrent processing)
				expectedCaptions := map[string]bool{
					"Property image 1": false,
					"Property image 2": false,
				}
				
				for _, photo := range photos {
					if _, exists := expectedCaptions[photo.Caption]; exists {
						expectedCaptions[photo.Caption] = true
					} else {
						t.Errorf("Unexpected caption '%s'", photo.Caption)
					}
					
					if !strings.Contains(photo.LocalURL, "prop123") {
						t.Errorf("Expected local URL to contain property ID, got '%s'", photo.LocalURL)
					}
				}
				
				// Verify all expected captions were found
				for caption, found := range expectedCaptions {
					if !found {
						t.Errorf("Expected caption '%s' not found", caption)
					}
				}
			},
		},
		{
			name:       "empty image URLs",
			imageURLs:  []string{},
			propertyID: "prop456",
			setupServer: func() *httptest.Server {
				return nil
			},
			expectError: false,
			verifyResult: func(t *testing.T, photos models.PhotoList) {
				if len(photos) != 0 {
					t.Errorf("Expected 0 photos, got %d", len(photos))
				}
			},
		},
		{
			name:       "server returns 404 for image",
			imageURLs:  []string{"/nonexistent.jpg"},
			propertyID: "prop789",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectError: true,
			errorMsg:    "some images failed to download",
			verifyResult: func(t *testing.T, photos models.PhotoList) {
				// May have partial results
			},
		},
	}

	// Create temporary directory for images
	tempDir, err := os.MkdirTemp("", "simplyrets_images_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			service := NewSimplyRETSService(mockRepo)
			service.imagesDir = tempDir

			var imageURLs []string
			if tt.setupServer != nil {
				server := tt.setupServer()
				if server != nil {
					defer server.Close()
					// Convert relative URLs to absolute
					for _, url := range tt.imageURLs {
						imageURLs = append(imageURLs, server.URL+url)
					}
				}
			} else {
				imageURLs = tt.imageURLs
			}

			ctx := context.Background()
			photos, err := service.downloadImages(ctx, imageURLs, tt.propertyID)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			tt.verifyResult(t, photos)
		})
	}
}

func TestSimplyRETSService_downloadImage(t *testing.T) {
	tests := []struct {
		name         string
		imageURL     string
		propertyID   string
		index        int
		setupServer  func() *httptest.Server
		expectError  bool
		errorMsg     string
		verifyResult func(t *testing.T, localPath string)
	}{
		{
			name:       "successful JPEG download",
			imageURL:   "/test.jpg",
			propertyID: "prop123",
			index:      0,
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "image/jpeg")
					w.Write([]byte("fake jpeg data"))
				}))
			},
			expectError: false,
			verifyResult: func(t *testing.T, localPath string) {
				if !strings.Contains(localPath, "/images/prop123_0.jpg") {
					t.Errorf("Expected path to contain '/images/prop123_0.jpg', got '%s'", localPath)
				}
			},
		},
		{
			name:       "successful PNG download",
			imageURL:   "/test.png",
			propertyID: "prop456",
			index:      1,
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "image/png")
					w.Write([]byte("fake png data"))
				}))
			},
			expectError: false,
			verifyResult: func(t *testing.T, localPath string) {
				if !strings.Contains(localPath, "/images/prop456_1.png") {
					t.Errorf("Expected path to contain '/images/prop456_1.png', got '%s'", localPath)
				}
			},
		},
		{
			name:       "server returns 500 error",
			imageURL:   "/error.jpg",
			propertyID: "prop789",
			index:      0,
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectError: true,
			errorMsg:    "image download returned status 500",
			verifyResult: func(t *testing.T, localPath string) {
				// Should not be called
			},
		},
	}

	// Create temporary directory for images
	tempDir, err := os.MkdirTemp("", "simplyrets_single_image_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			service := NewSimplyRETSService(mockRepo)
			service.imagesDir = tempDir

			server := tt.setupServer()
			defer server.Close()

			imageURL := server.URL + tt.imageURL
			ctx := context.Background()
			localPath, err := service.downloadImage(ctx, imageURL, tt.propertyID, tt.index)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				tt.verifyResult(t, localPath)

				// Verify file was actually created
				fullPath := filepath.Join(tempDir, filepath.Base(localPath))
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Error("Image file was not created")
				}
			}
		})
	}
}

func TestSimplyRETSService_convertToProperty(t *testing.T) {
	tests := []struct {
		name           string
		simplyProperty models.SimplyRETSProperty
		photos         models.PhotoList
		verifyResult   func(t *testing.T, property models.Property)
	}{
		{
			name: "convert property with all fields",
			simplyProperty: models.SimplyRETSProperty{
				ListingID: "12345",
				MLSNumber: "MLS12345",
				Address: models.SimplyRETSAddress{
					Full:         "123 Main Street, Anytown, ST 12345",
					StreetNumber: "123",
					StreetName:   "Main Street",
				},
				ListPrice: 450000.0,
				Property: models.SimplyRETSPropertyDetails{
					PropertyType: "Single Family Residential",
					Bedrooms:     3,
					Bathrooms:    2,
					Area:         1800,
					YearBuilt:    2010,
					LotSize:      "0.25 acres",
				},
				Remarks: "Beautiful family home with modern amenities",
			},
			photos: models.PhotoList{
				{URL: "http://example.com/photo1.jpg", LocalURL: "/images/12345_0.jpg"},
				{URL: "http://example.com/photo2.jpg", LocalURL: "/images/12345_1.jpg"},
			},
			verifyResult: func(t *testing.T, property models.Property) {
				expectedName := "123 Main Street"
				if property.Name != expectedName {
					t.Errorf("Expected name '%s', got '%s'", expectedName, property.Name)
				}
				if property.Location != "123 Main Street, Anytown, ST 12345" {
					t.Errorf("Expected location '123 Main Street, Anytown, ST 12345', got '%s'", property.Location)
				}
				if property.Price != 450000.0 {
					t.Errorf("Expected price 450000.0, got %f", property.Price)
				}
				if !property.Description.Valid || property.Description.String != "Beautiful family home with modern amenities" {
					t.Errorf("Expected description to be valid with correct value, got %+v", property.Description)
				}
				if len(property.Photos) != 2 {
					t.Errorf("Expected 2 photos, got %d", len(property.Photos))
				}
				if !property.ExternalID.Valid || property.ExternalID.String != "12345" {
					t.Errorf("Expected external ID to be '12345', got %+v", property.ExternalID)
				}
				if !property.MLSNumber.Valid || property.MLSNumber.String != "MLS12345" {
					t.Errorf("Expected MLS number to be 'MLS12345', got %+v", property.MLSNumber)
				}
				if !property.PropertyType.Valid || property.PropertyType.String != "Single Family Residential" {
					t.Errorf("Expected property type to be 'Single Family Residential', got %+v", property.PropertyType)
				}
				if !property.Bedrooms.Valid || property.Bedrooms.Int32 != 3 {
					t.Errorf("Expected bedrooms to be 3, got %+v", property.Bedrooms)
				}
				if !property.Bathrooms.Valid || property.Bathrooms.Int32 != 2 {
					t.Errorf("Expected bathrooms to be 2, got %+v", property.Bathrooms)
				}
				if !property.SquareFeet.Valid || property.SquareFeet.Int32 != 1800 {
					t.Errorf("Expected square feet to be 1800, got %+v", property.SquareFeet)
				}
				if !property.LotSize.Valid || property.LotSize.String != "0.25 acres" {
					t.Errorf("Expected lot size to be '0.25 acres', got %+v", property.LotSize)
				}
				if !property.YearBuilt.Valid || property.YearBuilt.Int32 != 2010 {
					t.Errorf("Expected year built to be 2010, got %+v", property.YearBuilt)
				}
			},
		},
		{
			name: "convert property with empty optional fields",
			simplyProperty: models.SimplyRETSProperty{
				ListingID: "67890",
				MLSNumber: "",
				Address: models.SimplyRETSAddress{
					Full:         "456 Oak Avenue, Springfield, ST 67890",
					StreetNumber: "456",
					StreetName:   "Oak Avenue",
				},
				ListPrice: 275000.0,
				Property: models.SimplyRETSPropertyDetails{
					PropertyType: "",
					Bedrooms:     0,
					Bathrooms:    0,
					Area:         0,
					YearBuilt:    0,
					LotSize:      "",
				},
				Remarks: "",
			},
			photos: models.PhotoList{},
			verifyResult: func(t *testing.T, property models.Property) {
				expectedName := "456 Oak Avenue"
				if property.Name != expectedName {
					t.Errorf("Expected name '%s', got '%s'", expectedName, property.Name)
				}
				if property.Description.Valid {
					t.Error("Expected description to be invalid (null)")
				}
				if property.MLSNumber.Valid {
					t.Error("Expected MLS number to be invalid (null)")
				}
				if property.PropertyType.Valid {
					t.Error("Expected property type to be invalid (null)")
				}
				if property.Bedrooms.Valid {
					t.Error("Expected bedrooms to be invalid (null)")
				}
				if property.Bathrooms.Valid {
					t.Error("Expected bathrooms to be invalid (null)")
				}
				if property.SquareFeet.Valid {
					t.Error("Expected square feet to be invalid (null)")
				}
				if property.LotSize.Valid {
					t.Error("Expected lot size to be invalid (null)")
				}
				if property.YearBuilt.Valid {
					t.Error("Expected year built to be invalid (null)")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			service := NewSimplyRETSService(mockRepo)

			property := service.convertToProperty(tt.simplyProperty, tt.photos)
			tt.verifyResult(t, property)
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("nullString", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected models.NullString
		}{
			{
				name:  "empty string",
				input: "",
				expected: models.NullString{
					NullString: sql.NullString{Valid: false},
				},
			},
			{
				name:  "non-empty string",
				input: "test",
				expected: models.NullString{
					NullString: sql.NullString{String: "test", Valid: true},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := nullString(tt.input)
				if result.Valid != tt.expected.Valid {
					t.Errorf("Expected Valid %t, got %t", tt.expected.Valid, result.Valid)
				}
				if result.Valid && result.String != tt.expected.String {
					t.Errorf("Expected String '%s', got '%s'", tt.expected.String, result.String)
				}
			})
		}
	})

	t.Run("nullInt32", func(t *testing.T) {
		tests := []struct {
			name     string
			input    int
			expected models.NullInt32
		}{
			{
				name:  "zero value",
				input: 0,
				expected: models.NullInt32{
					NullInt32: sql.NullInt32{Valid: false},
				},
			},
			{
				name:  "non-zero value",
				input: 42,
				expected: models.NullInt32{
					NullInt32: sql.NullInt32{Int32: 42, Valid: true},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := nullInt32(tt.input)
				if result.Valid != tt.expected.Valid {
					t.Errorf("Expected Valid %t, got %t", tt.expected.Valid, result.Valid)
				}
				if result.Valid && result.Int32 != tt.expected.Int32 {
					t.Errorf("Expected Int32 %d, got %d", tt.expected.Int32, result.Int32)
				}
			})
		}
	})
}

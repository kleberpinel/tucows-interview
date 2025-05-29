package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"real-estate-manager/backend/internal/models"
	"real-estate-manager/backend/internal/repository"
	"strings"
	"sync"
	"time"
)

type SimplyRETSService struct {
	propertyRepo repository.PropertyRepository
	client       *http.Client
	baseURL      string
	username     string
	password     string
	imagesDir    string
}

// ProcessingJob represents a property processing job
type ProcessingJob struct {
	ID           string
	Status       chan models.ProcessingStatus
	Cancel       context.CancelFunc
	StartTime    time.Time
	LastStatus   *models.ProcessingStatus
	CompletedAt  *time.Time
	mu           sync.RWMutex
}

// JobManager manages processing jobs
type JobManager struct {
	jobs map[string]*ProcessingJob
	mu   sync.RWMutex
}

const JobRetentionDuration = 5 * time.Minute // Keep completed jobs for 5 minutes

func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*ProcessingJob),
	}
}

func (jm *JobManager) AddJob(id string, job *ProcessingJob) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.jobs[id] = job
	log.Printf("Job %s added to manager (total jobs: %d)", id, len(jm.jobs))
}

func (jm *JobManager) GetJob(id string) (*ProcessingJob, bool) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	job, exists := jm.jobs[id]
	if !exists {
		log.Printf("Job %s not found in manager", id)
	}
	return job, exists
}

func (jm *JobManager) RemoveJob(id string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	if job, exists := jm.jobs[id]; exists {
		close(job.Status)
		delete(jm.jobs, id)
		log.Printf("Job %s removed from manager (remaining jobs: %d)", id, len(jm.jobs))
	} else {
		log.Printf("Attempted to remove non-existent job %s", id)
	}
}

func (jm *JobManager) MarkJobCompleted(id string, finalStatus models.ProcessingStatus) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	if job, exists := jm.jobs[id]; exists {
		job.mu.Lock()
		job.LastStatus = &finalStatus
		now := time.Now()
		job.CompletedAt = &now
		job.mu.Unlock()
		
		log.Printf("Job %s marked as completed with status: %s", id, finalStatus.Status)
		
		// Schedule cleanup after retention period
		go func() {
			log.Printf("Job %s cleanup scheduled in %v", id, JobRetentionDuration)
			time.Sleep(JobRetentionDuration)
			jm.CleanupJob(id)
		}()
	} else {
		log.Printf("Attempted to mark non-existent job %s as completed", id)
	}
}

func (jm *JobManager) CleanupJob(id string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	if job, exists := jm.jobs[id]; exists {
		job.mu.RLock()
		isCompleted := job.CompletedAt != nil
		completedTime := job.CompletedAt
		job.mu.RUnlock()
		
		if isCompleted && completedTime != nil && time.Since(*completedTime) >= JobRetentionDuration {
			close(job.Status)
			delete(jm.jobs, id)
			log.Printf("Job %s cleaned up after retention period (remaining jobs: %d)", id, len(jm.jobs))
		} else {
			log.Printf("Job %s cleanup skipped - not ready for cleanup", id)
		}
	} else {
		log.Printf("Job %s already removed during cleanup", id)
	}
}

var GlobalJobManager = NewJobManager()

func NewSimplyRETSService(propertyRepo repository.PropertyRepository) *SimplyRETSService {
	// Create images directory if it doesn't exist
	imagesDir := "./uploads/images"
	os.MkdirAll(imagesDir, 0755)

	return &SimplyRETSService{
		propertyRepo: propertyRepo,
		client:       &http.Client{Timeout: 30 * time.Second},
		baseURL:      "https://api.simplyrets.com",
		username:     "simplyrets",
		password:     "simplyrets",
		imagesDir:    imagesDir,
	}
}

// StartPropertyProcessing starts the property processing job
func (s *SimplyRETSService) StartPropertyProcessing(ctx context.Context, jobID string, limit int) error {
	log.Printf("Starting property processing job %s with limit %d", jobID, limit)
	
	// Create a cancellable context for this job
	jobCtx, cancel := context.WithCancel(ctx)
	
	// Create status channel
	statusChan := make(chan models.ProcessingStatus, 100)
	
	// Create and register the job
	job := &ProcessingJob{
		ID:          jobID,
		Status:      statusChan,
		Cancel:      cancel,
		StartTime:   time.Now(),
		LastStatus:  nil,
		CompletedAt: nil,
	}
	GlobalJobManager.AddJob(jobID, job)
	
	// Start processing in a goroutine
	go s.processProperties(jobCtx, jobID, statusChan, limit)
	
	log.Printf("Property processing job %s started successfully", jobID)
	return nil
}

// GetJobStatus returns the current status of a job
func (s *SimplyRETSService) GetJobStatus(jobID string) (*models.ProcessingStatus, bool) {
	job, exists := GlobalJobManager.GetJob(jobID)
	if !exists {
		log.Printf("GetJobStatus: Job %s not found", jobID)
		return nil, false
	}
	
	job.mu.RLock()
	defer job.mu.RUnlock()
	
	// If job is completed, return the final status
	if job.LastStatus != nil {
		log.Printf("GetJobStatus: Returning completed status for job %s: %s", jobID, job.LastStatus.Status)
		return job.LastStatus, true
	}
	
	// For running jobs, try to get the latest status without blocking
	// Use a non-blocking select to avoid consuming the status update
	select {
	case status := <-job.Status:
		// Update the last status and put it back for other potential readers
		job.mu.RUnlock()
		job.mu.Lock()
		job.LastStatus = &status
		job.mu.Unlock()
		job.mu.RLock()
		
		log.Printf("GetJobStatus: Updated status for job %s: %s (processed: %d/%d)", jobID, status.Status, status.ProcessedCount, status.TotalProperties)
		
		// Try to put the status back (non-blocking)
		select {
		case job.Status <- status:
		default:
			// Channel full, that's OK
		}
		
		return &status, true
	default:
		// Return a basic status if no update is available
		log.Printf("GetJobStatus: No status update available for job %s, returning default running status", jobID)
		return &models.ProcessingStatus{
			Status:    "running",
			StartedAt: job.StartTime,
		}, true
	}
}

// CancelJob cancels a running job
func (s *SimplyRETSService) CancelJob(jobID string) bool {
	log.Printf("Attempting to cancel job %s", jobID)
	job, exists := GlobalJobManager.GetJob(jobID)
	if !exists {
		log.Printf("Cannot cancel job %s: job not found", jobID)
		return false
	}
	
	job.Cancel()
	GlobalJobManager.RemoveJob(jobID)
	log.Printf("Job %s cancelled successfully", jobID)
	return true
}

// processProperties is the main processing function that runs in a goroutine
func (s *SimplyRETSService) processProperties(ctx context.Context, jobID string, statusChan chan models.ProcessingStatus, limit int) {
	log.Printf("processProperties: Starting job %s with limit %d", jobID, limit)
	
	// Send initial status
	status := models.ProcessingStatus{
		Status:          "running",
		TotalProperties: 0,
		ProcessedCount:  0,
		FailedCount:     0,
		StartedAt:       time.Now(),
	}
	
	log.Printf("processProperties: Sending initial status for job %s", jobID)
	select {
	case statusChan <- status:
		log.Printf("processProperties: Initial status sent successfully for job %s", jobID)
	case <-ctx.Done():
		log.Printf("processProperties: Context cancelled before sending initial status for job %s", jobID)
		return
	}
	
	// Fetch properties from SimplyRETS
	log.Printf("processProperties: Fetching properties from SimplyRETS for job %s (limit: %d)", jobID, limit)
	properties, err := s.fetchProperties(ctx, limit)
	if err != nil {
		log.Printf("processProperties: Failed to fetch properties for job %s: %v", jobID, err)
		status.Status = "failed"
		status.ErrorMessage = err.Error()
		completedAt := time.Now()
		status.CompletedAt = &completedAt
		statusChan <- status
		GlobalJobManager.MarkJobCompleted(jobID, status)
		return
	}
	
	log.Printf("processProperties: Successfully fetched %d properties for job %s", len(properties), jobID)
	status.TotalProperties = len(properties)
	statusChan <- status
	
	// Process properties in batches of 10
	batchSize := 10
	log.Printf("processProperties: Starting batch processing for job %s (%d properties, batch size: %d)", jobID, len(properties), batchSize)
	
	for i := 0; i < len(properties); i += batchSize {
		select {
		case <-ctx.Done():
			log.Printf("processProperties: Context cancelled during processing for job %s", jobID)
			status.Status = "cancelled"
			completedAt := time.Now()
			status.CompletedAt = &completedAt
			statusChan <- status
			GlobalJobManager.MarkJobCompleted(jobID, status)
			return
		default:
		}
		
		end := i + batchSize
		if end > len(properties) {
			end = len(properties)
		}
		
		log.Printf("processProperties: Processing batch %d-%d for job %s", i+1, end, jobID)
		
		batch := properties[i:end]
		s.processBatch(ctx, batch, statusChan, &status)
		log.Printf("processProperties: Completed batch %d-%d for job %s (total processed: %d, failed: %d)", i+1, end, jobID, status.ProcessedCount, status.FailedCount)
	}
	
	// Send final status
	log.Printf("processProperties: Job %s completed successfully. Total: %d, Processed: %d, Failed: %d", jobID, status.TotalProperties, status.ProcessedCount, status.FailedCount)
	status.Status = "completed"
	completedAt := time.Now()
	status.CompletedAt = &completedAt
	statusChan <- status
	GlobalJobManager.MarkJobCompleted(jobID, status)
}

// fetchProperties fetches properties from SimplyRETS API
func (s *SimplyRETSService) fetchProperties(ctx context.Context, limit int) ([]models.SimplyRETSProperty, error) {
	url := fmt.Sprintf("%s/properties?limit=%d", s.baseURL, limit)
	log.Printf("fetchProperties: Making request to %s", url)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("fetchProperties: Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Accept", "application/json")
	
	log.Printf("fetchProperties: Sending request to SimplyRETS API")
	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("fetchProperties: Request failed: %v", err)
		return nil, fmt.Errorf("failed to fetch properties: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("fetchProperties: Received non-200 status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	log.Printf("fetchProperties: Successfully received response, decoding JSON")
	var properties []models.SimplyRETSProperty
	if err := json.NewDecoder(resp.Body).Decode(&properties); err != nil {
		log.Printf("fetchProperties: Failed to decode JSON response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	log.Printf("fetchProperties: Successfully fetched and decoded %d properties", len(properties))
	return properties, nil
}

// processBatch processes a batch of properties
func (s *SimplyRETSService) processBatch(ctx context.Context, batch []models.SimplyRETSProperty, statusChan chan models.ProcessingStatus, status *models.ProcessingStatus) {
	log.Printf("processBatch: Processing batch of %d properties", len(batch))
	var wg sync.WaitGroup
	results := make(chan error, len(batch))
	
	// Process each property in the batch concurrently
	for i, prop := range batch {
		wg.Add(1)
		go func(idx int, property models.SimplyRETSProperty) {
			defer wg.Done()
			
			select {
			case <-ctx.Done():
				log.Printf("processBatch: Context cancelled while processing property %d in batch", idx+1)
				results <- ctx.Err()
				return
			default:
			}
			
			log.Printf("processBatch: Processing property %d (MLS: %s)", idx+1, property.MLSNumber.String())
			err := s.processProperty(ctx, property)
			if err != nil {
				log.Printf("processBatch: Failed to process property %d (MLS: %s): %v", idx+1, property.MLSNumber.String(), err)
			} else {
				log.Printf("processBatch: Successfully processed property %d (MLS: %s)", idx+1, property.MLSNumber.String())
			}
			results <- err
		}(i, prop)
	}
	
	// Wait for all goroutines to complete
	log.Printf("processBatch: Waiting for all %d properties to complete processing", len(batch))
	wg.Wait()
	close(results)
	
	// Collect results and update status
	for err := range results {
		if err != nil {
			status.FailedCount++
		} else {
			status.ProcessedCount++
		}
	}
	
	// Send updated status
	select {
	case statusChan <- *status:
	case <-ctx.Done():
		return
	}
}

// processProperty processes a single property
func (s *SimplyRETSService) processProperty(ctx context.Context, simplyProperty models.SimplyRETSProperty) error {
	// Download images in parallel
	photos, err := s.downloadImages(ctx, simplyProperty.Photos, simplyProperty.ListingID)
	if err != nil {
		return fmt.Errorf("failed to download images for property %s: %w", simplyProperty.ListingID, err)
	}
	
	// Convert SimplyRETS property to our Property model
	property := s.convertToProperty(simplyProperty, photos)
	
	// Save to database
	if err := s.propertyRepo.Create(ctx, &property); err != nil {
		return fmt.Errorf("failed to save property %s: %w", simplyProperty.ListingID, err)
	}
	
	return nil
}

// downloadImages downloads property images in parallel
func (s *SimplyRETSService) downloadImages(ctx context.Context, imageURLs []string, propertyID string) (models.PhotoList, error) {
	if len(imageURLs) == 0 {
		return models.PhotoList{}, nil
	}
	
	var wg sync.WaitGroup
	photosChan := make(chan models.Photo, len(imageURLs))
	errorsChan := make(chan error, len(imageURLs))
	
	// Download each image concurrently
	for i, url := range imageURLs {
		wg.Add(1)
		go func(imageURL string, index int) {
			defer wg.Done()
			
			select {
			case <-ctx.Done():
				errorsChan <- ctx.Err()
				return
			default:
			}
			
			localPath, err := s.downloadImage(ctx, imageURL, propertyID, index)
			if err != nil {
				errorsChan <- err
				return
			}
			
			photo := models.Photo{
				URL:      imageURL,
				LocalURL: localPath,
				Caption:  fmt.Sprintf("Property image %d", index+1),
			}
			
			photosChan <- photo
		}(url, i)
	}
	
	// Wait for all downloads to complete
	wg.Wait()
	close(photosChan)
	close(errorsChan)
	
	// Collect results
	var photos models.PhotoList
	for photo := range photosChan {
		photos = append(photos, photo)
	}
	
	// Check for errors
	var errors []string
	for err := range errorsChan {
		errors = append(errors, err.Error())
	}
	
	if len(errors) > 0 {
		return photos, fmt.Errorf("some images failed to download: %s", strings.Join(errors, "; "))
	}
	
	return photos, nil
}

// downloadImage downloads a single image
func (s *SimplyRETSService) downloadImage(ctx context.Context, imageURL, propertyID string, index int) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create image request: %w", err)
	}
	
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image download returned status %d", resp.StatusCode)
	}
	
	// Generate filename
	ext := ".jpg"
	if strings.Contains(resp.Header.Get("Content-Type"), "png") {
		ext = ".png"
	}
	filename := fmt.Sprintf("%s_%d%s", propertyID, index, ext)
	filePath := filepath.Join(s.imagesDir, filename)
	
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()
	
	// Copy image data
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}
	
	// Return relative path for API access
	return fmt.Sprintf("/images/%s", filename), nil
}

// Helper functions for creating custom null types
func nullString(s string) models.NullString {
	if s == "" {
		return models.NullString{NullString: sql.NullString{Valid: false}}
	}
	return models.NullString{NullString: sql.NullString{String: s, Valid: true}}
}

func nullInt32(i int) models.NullInt32 {
	if i == 0 {
		return models.NullInt32{NullInt32: sql.NullInt32{Valid: false}}
	}
	return models.NullInt32{NullInt32: sql.NullInt32{Int32: int32(i), Valid: true}}
}

// convertToProperty converts SimplyRETS property to our Property model
func (s *SimplyRETSService) convertToProperty(simplyProperty models.SimplyRETSProperty, photos models.PhotoList) models.Property {
	return models.Property{
		Name:         fmt.Sprintf("%s %s", simplyProperty.Address.StreetNumber.String(), simplyProperty.Address.StreetName),
		Location:     simplyProperty.Address.Full,
		Price:        simplyProperty.ListPrice,
		Description:  nullString(simplyProperty.Remarks),
		Photos:       photos,
		ExternalID:   nullString(simplyProperty.ListingID),
		MLSNumber:    nullString(simplyProperty.MLSNumber.String()),
		PropertyType: nullString(simplyProperty.Property.PropertyType),
		Bedrooms:     nullInt32(simplyProperty.Property.Bedrooms),
		Bathrooms:    nullInt32(simplyProperty.Property.Bathrooms),
		SquareFeet:   nullInt32(simplyProperty.Property.Area),
		LotSize:      nullString(simplyProperty.Property.LotSize),
		YearBuilt:    nullInt32(simplyProperty.Property.YearBuilt),
	}
}

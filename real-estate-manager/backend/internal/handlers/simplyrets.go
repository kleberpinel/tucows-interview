package handlers

import (
	"context"
	"fmt"
	"net/http"
	"real-estate-manager/backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SimplyRETSHandler struct {
	simplyRETSService *services.SimplyRETSService
}

func NewSimplyRETSHandler(simplyRETSService *services.SimplyRETSService) *SimplyRETSHandler {
	return &SimplyRETSHandler{
		simplyRETSService: simplyRETSService,
	}
}

// StartProcessing starts the property processing job
func (h *SimplyRETSHandler) StartProcessing(c *gin.Context) {
	var request struct {
		Limit int `json:"limit"`
	}
	
	// Default limit to 50 if not provided
	request.Limit = 50
	
	if err := c.ShouldBindJSON(&request); err != nil {
		// If binding fails, use query parameter or default
		if limitStr := c.Query("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				request.Limit = limit
			}
		}
	}
	
	// Validate limit
	if request.Limit <= 0 || request.Limit > 500 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Limit must be between 1 and 500",
		})
		return
	}
	
	// Generate unique job ID
	jobID := uuid.New().String()
	
	// Start processing with a background context instead of request context
	// This prevents the job from being cancelled when the HTTP request completes
	err := h.simplyRETSService.StartPropertyProcessing(context.Background(), jobID, request.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to start processing: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusAccepted, gin.H{
		"job_id":    jobID,
		"message":   "Property processing started",
		"limit":     request.Limit,
		"started_at": time.Now(),
	})
}

// GetJobStatus returns the status of a processing job
func (h *SimplyRETSHandler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}
	
	status, exists := h.simplyRETSService.GetJobStatus(jobID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, status)
}

// CancelJob cancels a running processing job
func (h *SimplyRETSHandler) CancelJob(c *gin.Context) {
	jobID := c.Param("jobId")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}
	
	cancelled := h.simplyRETSService.CancelJob(jobID)
	if !cancelled {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found or already completed",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Job cancelled successfully",
		"job_id":  jobID,
	})
}

// GetProcessingHistory returns a summary of processing activities
func (h *SimplyRETSHandler) GetProcessingHistory(c *gin.Context) {
	// This would typically come from a database table storing job history
	// For now, we'll return a simple response
	c.JSON(http.StatusOK, gin.H{
		"message": "Processing history endpoint - to be implemented with persistent storage",
	})
}

// HealthCheck returns the health status of the SimplyRETS service
func (h *SimplyRETSHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "SimplyRETS Integration",
		"timestamp": time.Now(),
	})
}

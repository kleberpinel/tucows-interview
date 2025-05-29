package handlers

import (
	"net/http"
	"real-estate-manager/backend/internal/models"
	services "real-estate-manager/backend/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PropertyHandler struct {
	Service *services.PropertyService
}

// NewPropertyHandler creates a new PropertyHandler instance
func NewPropertyHandler(service *services.PropertyService) *PropertyHandler {
	return &PropertyHandler{
		Service: service,
	}
}

func (h *PropertyHandler) CreateProperty(c *gin.Context) {
	var property models.Property
	if err := c.ShouldBindJSON(&property); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.Service.CreateProperty(c.Request.Context(), &property)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, property)
}

func (h *PropertyHandler) GetProperties(c *gin.Context) {
	properties, err := h.Service.GetAllProperties(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, properties)
}

func (h *PropertyHandler) GetProperty(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	property, err := h.Service.GetProperty(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, property)
}

func (h *PropertyHandler) UpdateProperty(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	var property models.Property
	if err := c.ShouldBindJSON(&property); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	property.ID = id
	err = h.Service.UpdateProperty(c.Request.Context(), &property)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, property)
}

func (h *PropertyHandler) DeleteProperty(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	if err := h.Service.DeleteProperty(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Property deleted successfully"})
}
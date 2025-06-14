package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hass-security/hass-security/webapp/backend/pkg/database"
	"github.com/hass-security/hass-security/webapp/backend/pkg/thresholds"
	"github.com/sirupsen/logrus"
)

func GetDeviceDetails(c *gin.Context) {
	logger := c.MustGet("LOGGER").(*logrus.Entry)
	deviceRepo := c.MustGet("DEVICE_REPOSITORY").(database.DeviceRepo)

	device, err := deviceRepo.GetDeviceDetails(c, c.Param("wwn"))
	if err != nil {
		logger.Errorln("An error occurred while retrieving device details", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	durationKey, exists := c.GetQuery("duration_key")
	if !exists {
		durationKey = "forever"
	}

	smartResults, err := deviceRepo.GetSmartAttributeHistory(c, c.Param("wwn"), durationKey, 0, 0, nil)
	if err != nil {
		logger.Errorln("An error occurred while retrieving device smart results", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	var deviceMetadata interface{}
	if device.IsAta() {
		deviceMetadata = thresholds.AtaMetadata
	} else if device.IsNvme() {
		deviceMetadata = thresholds.NmveMetadata
	} else if device.IsScsi() {
		deviceMetadata = thresholds.ScsiMetadata
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": map[string]interface{}{"device": device, "smart_results": smartResults}, "metadata": deviceMetadata})
}

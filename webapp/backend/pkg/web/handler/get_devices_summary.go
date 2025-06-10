package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hass-security/hass-security/webapp/backend/pkg/database"
	"github.com/sirupsen/logrus"
	"net/http"
)

func GetDevicesSummary(c *gin.Context) {
	logger := c.MustGet("LOGGER").(*logrus.Entry)
	deviceRepo := c.MustGet("DEVICE_REPOSITORY").(database.DeviceRepo)

	summary, err := deviceRepo.GetSummary(c)
	if err != nil {
		logger.Errorln("An error occurred while retrieving device summary", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	//this must match DeviceSummaryWrapper (webapp/backend/pkg/models/device_summary.go)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"summary": summary,
			//"temperature": tem
		},
	})
}

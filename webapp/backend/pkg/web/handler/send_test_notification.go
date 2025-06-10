package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hass-security/hass-security/webapp/backend/pkg"
	"github.com/hass-security/hass-security/webapp/backend/pkg/config"
	"github.com/hass-security/hass-security/webapp/backend/pkg/models"
	"github.com/hass-security/hass-security/webapp/backend/pkg/notify"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Send test notification
func SendTestNotification(c *gin.Context) {
	appConfig := c.MustGet("CONFIG").(config.Interface)
	logger := c.MustGet("LOGGER").(*logrus.Entry)

	testNotify := notify.New(
		logger,
		appConfig,
		models.Device{
			SerialNumber: "FAKEWDDJ324KSO",
			DeviceType:   pkg.DeviceProtocolAta,
			DeviceName:   "/dev/sda",
		},
		true,
	)
	err := testNotify.Send()
	if err != nil {
		logger.Errorln("An error occurred while sending test notification", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"errors":  []string{err.Error()},
		})
	} else {
		c.JSON(http.StatusOK, models.DeviceWrapper{
			Success: true,
		})
	}
}

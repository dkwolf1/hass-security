package detect_test

import (
	"github.com/hass-security/hass-security/collector/pkg/detect"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDevicePrefix(t *testing.T) {
	//setup

	//test

	//assert
	require.Equal(t, "/dev/", detect.DevicePrefix())
}

package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Diniboy1123/usque/internal"
)

// Heartbeat sends a lightweight GET request to the device registration endpoint
// to update the device's "last seen" timestamp on Cloudflare's servers.
//
// This mimics the behavior of the official WARP client's periodic heartbeat.
// Unlike enroll, it does not regenerate keys or download new config — it simply
// touches the API to signal that the device is still active.
//
// Parameters:
//   - deviceId: string - The device registration ID
//   - deviceToken: string - The device registration access token
//
// Returns:
//   - error: An error if the heartbeat fails
func Heartbeat(deviceId string, deviceToken string) error {
	req, err := http.NewRequest(http.MethodGet, internal.ApiUrl+"/"+internal.ApiVersion+"/reg/"+deviceId, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	for k, v := range internal.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+deviceToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode, string(body))
	}

	_, _ = io.ReadAll(resp.Body)

	return nil
}

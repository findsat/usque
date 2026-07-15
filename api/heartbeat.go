package api

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Diniboy1123/usque/internal"
)

// Heartbeat sends a lightweight PATCH request to the device registration endpoint
// to update the device's "last seen" timestamp on Cloudflare's servers.
//
// Unlike enroll, this does NOT save the config or restart the tunnel.
// It re-sends the existing public key to refresh the registration timestamp.
//
// Parameters:
//   - deviceId: string - The device registration ID
//   - deviceToken: string - The device registration access token
//   - privKeyPemBase64: string - Base64 encoded private key from config
//
// Returns:
//   - error: An error if the heartbeat fails
func Heartbeat(deviceId string, deviceToken string, privKeyPemBase64 string) error {
	// Decode the private key from config
	privKeyBytes, err := base64.StdEncoding.DecodeString(privKeyPemBase64)
	if err != nil {
		return fmt.Errorf("failed to decode private key: %v", err)
	}

	privKey, err := x509.ParseECPrivateKey(privKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	// Get the existing public key (same key, no regeneration)
	pubKey, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}

	// Send PATCH with the same key — this refreshes the registration timestamp
	// without changing anything
	deviceUpdate := struct {
		Key     string `json:"key"`
		KeyType string `json:"key_type"`
		TunType string `json:"tun_type"`
	}{
		Key:     base64.StdEncoding.EncodeToString(pubKey),
		KeyType: internal.KeyTypeMasque,
		TunType: internal.TunTypeMasque,
	}

	jsonData, err := json.Marshal(deviceUpdate)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %v", err)
	}

	req, err := http.NewRequest(http.MethodPatch, internal.ApiUrl+"/"+internal.ApiVersion+"/reg/"+deviceId, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	for k, v := range internal.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+deviceToken)
	req.Header.Set("Content-Type", "application/json")

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

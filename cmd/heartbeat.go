package cmd

import (
	"log"
	"time"

	"github.com/Diniboy1123/usque/api"
	"github.com/Diniboy1123/usque/config"
	"github.com/spf13/cobra"
)

var heartbeatCmd = &cobra.Command{
	Use:   "heartbeat",
	Short: "Send a heartbeat to keep the device alive in Zero Trust dashboard",
	Long: "Sends a lightweight API request to Cloudflare to update the device's last-seen timestamp.\n" +
		"Can run as a daemon with --interval to periodically send heartbeats.\n" +
		"Unlike enroll, this does not regenerate keys or disrupt the tunnel.",
	Run: func(cmd *cobra.Command, args []string) {
		if !config.ConfigLoaded {
			cmd.Println("Config not loaded. Please register first.")
			return
		}

		interval, err := cmd.Flags().GetDuration("interval")
		if err != nil {
			log.Fatalf("Failed to get interval: %v", err)
		}

		if interval == 0 {
			// Single heartbeat mode
			if err := api.Heartbeat(config.AppConfig.ID, config.AppConfig.AccessToken); err != nil {
				log.Fatalf("Heartbeat failed: %v", err)
			}
			log.Println("Heartbeat successful")
			return
		}

		// Daemon mode — send heartbeat periodically
		log.Printf("Starting heartbeat daemon, interval: %s", interval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Send immediately on start
		if err := api.Heartbeat(config.AppConfig.ID, config.AppConfig.AccessToken); err != nil {
			log.Printf("Heartbeat failed: %v", err)
		} else {
			log.Println("Heartbeat successful")
		}

		for range ticker.C {
			if err := api.Heartbeat(config.AppConfig.ID, config.AppConfig.AccessToken); err != nil {
				log.Printf("Heartbeat failed: %v", err)
			}
		}
	},
}

func init() {
	heartbeatCmd.Flags().DurationP("interval", "i", 0, "Run as daemon, sending heartbeat at specified interval (e.g. 60s, 5m). Default: single heartbeat")
	rootCmd.AddCommand(heartbeatCmd)
}

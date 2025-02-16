package cmd

import (
	"blink-liveview-websocket/liveview"

	"github.com/spf13/cobra"
)

var liveviewCmd = &cobra.Command{
	Use:   "liveview",
	Short: "Start a liveview stream directly with specified credentials",
	Long: `The liveview command exposes a direct method to start a liveview stream
without the need for logging in or utilizing the WebSocket server.

You can use this command if you already have all of the connection credentials. 
If you do not have all of the required information, use the account command instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		accountId, _ := cmd.Flags().GetInt("account-id")
		networkId, _ := cmd.Flags().GetInt("network-id")
		cameraId, _ := cmd.Flags().GetInt("camera-id")

		liveview.Run(
			cmd.Flag("region").Value.String(),
			cmd.Flag("token").Value.String(),
			cmd.Flag("device-type").Value.String(),
			accountId,
			networkId,
			cameraId,
		)
	},
}

func init() {
	rootCmd.AddCommand(liveviewCmd)

	liveviewCmd.Flags().StringP("region", "r", "", "The Blink API subdomain/region to use (e.g. u011)")
	liveviewCmd.MarkFlagRequired("region")
	liveviewCmd.Flags().StringP("token", "t", "", "The Blink API token to use for authentication")
	liveviewCmd.MarkFlagRequired("token")
	liveviewCmd.Flags().StringP("device-type", "d", "", "The Blink device type (e.g. owl, doorbell, etc)")
	liveviewCmd.MarkFlagRequired("device-type")
	liveviewCmd.Flags().IntP("account-id", "a", 0, "The Blink account ID")
	liveviewCmd.MarkFlagRequired("account-id")
	liveviewCmd.Flags().IntP("network-id", "n", 0, "The Blink network ID")
	liveviewCmd.MarkFlagRequired("network-id")
	liveviewCmd.Flags().IntP("camera-id", "c", 0, "The Blink camera ID")
	liveviewCmd.MarkFlagRequired("camera-id")
}

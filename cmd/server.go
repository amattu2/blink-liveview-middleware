package cmd

import (
	"blink-liveview-websocket/server"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a WebSocket middleware server to proxy liveview streams to clients",
	Long: `This command starts a WebSocket server that will proxy the liveview streams to the clients.
Each client acts as an independent subscriber to the streams,
which means many clients can connect and view their own liveview stream simultaneously.`,
	Run: func(cmd *cobra.Command, args []string) {
		server.Run(cmd.Flag("address").Value.String(), cmd.Flag("env").Value.String())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("address", "a", ":8080", "HTTP server address")
	serverCmd.Flags().StringP("env", "e", "production", "Environment (development, production)")
}

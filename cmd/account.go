package cmd

import (
	"blink-liveview-websocket/account"
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Start a liveview stream by logging in with your Blink account and selecting a camera",
	Long: `This command will authenticate with your Blink account, fetch a list of available cameras,
and start a liveview stream from the selected camera.

Use this command if you want to start a liveview stream, but do not have the
connection credentials already.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			os.Exit(1)
		}
		pass := string(passwordBytes)
		fmt.Println()

		account.Run(cmd.Flag("email").Value.String(), pass)
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)

	accountCmd.Flags().StringP("email", "e", "", "Blink account email address")
	accountCmd.MarkFlagRequired("email")
}

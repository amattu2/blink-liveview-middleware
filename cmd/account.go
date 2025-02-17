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

Additionally, you can provide a token and region to bypass the login process.

Use this command if you want to start a liveview stream, but do not have the
full connection credentials already.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("email").Value.String() != "" {
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				os.Exit(1)
			}
			pass := string(passwordBytes)
			fmt.Println()

			account.RunWithCredentials(cmd.Flag("email").Value.String(), pass)
			return
		}

		accountId, _ := cmd.Flags().GetInt("account-id")
		account.Run(cmd.Flag("token").Value.String(), accountId, cmd.Flag("region").Value.String())
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)

	accountCmd.Flags().StringP("email", "e", "", "Blink account email address")

	accountCmd.Flags().StringP("token", "t", "", "Blink auth token")
	accountCmd.Flags().IntP("account-id", "a", 0, "Blink account ID")
	accountCmd.Flags().StringP("region", "r", "", "Blink API region")
	accountCmd.MarkFlagsRequiredTogether("token", "account-id", "region")

	accountCmd.MarkFlagsMutuallyExclusive("email", "token")
	accountCmd.MarkFlagsOneRequired("email", "token")
}

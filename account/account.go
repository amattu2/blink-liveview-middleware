package account

import (
	"blink-liveview-websocket/common"
	"fmt"
	"log"
	"os"
	"syscall"

	"golang.org/x/term"
)

func Run(email *string, password *string) {
	if *email == "" {
		fmt.Fprintf(os.Stderr, "No email parameter provided. Please specify via --email=<email address>\n")
		os.Exit(1)
	}

	if *password == "" {
		fmt.Fprintf(os.Stderr, "No password provided.\n")
		os.Exit(1)
	}

	resp, err := common.Login(*email, *password)
	if err != nil {
		log.Println("error logging in", err)
		os.Exit(1)
	}

	if resp.Account.ClientVerificationRequired {
		log.Println("Client verification is required. A SMS code has been sent to your phone.")

		fmt.Print("Code: ")
		codeBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			os.Exit(1)
		}
		code := string(codeBytes)
		fmt.Println()
		fmt.Printf("Code: %s\n", code) // TODO: Call 2FA with code
	}

	// TODO: Fetch all devices and prompt user to select one

	// TODO: Begin live view for selected device
}

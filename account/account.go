package account

import (
	"blink-liveview-websocket/common"
	"fmt"
	"log"
	"os"
	"syscall"

	"golang.org/x/term"
)

// VerifyClient prompts the user for a code and verifies the client with the Blink API
//
// resp: the login response to use for verification
//
// Example: VerifyClient(LoginResponse{}) = nil
func verifyClient(resp common.LoginResponse) error {
	fmt.Print("Code: ")
	codeBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		os.Exit(1)
	}
	code := string(codeBytes)
	fmt.Println()

	baseUrl := common.GetApiUrl(resp.Account.Tier)
	verifyUrl := fmt.Sprintf("%s/api/v4/account/%d/client/%d/pin/verify", baseUrl, resp.Account.AccountId, resp.Account.ClientId)
	if err := common.VerifyPin(verifyUrl, resp.Auth.Token, code); err != nil {
		return err
	}

	return nil
}

func Run(email string, password string) {
	if email == "" {
		fmt.Fprintf(os.Stderr, "No email parameter provided. Please specify via --email=<email address>\n")
		os.Exit(1)
	}

	if password == "" {
		fmt.Fprintf(os.Stderr, "No password provided.\n")
		os.Exit(1)
	}

	fingerprint, err := common.GetFingerprint("")
	if err != nil {
		log.Println("error getting fingerprint", err)
		os.Exit(1)
	}

	resp, err := common.Login(email, password, fingerprint)
	if err != nil {
		log.Println("error logging in", err)
		os.Exit(1)
	}

	if resp.Account.ClientVerificationRequired {
		log.Println("Client verification is required. A SMS code has been sent to your phone.")
		if err := verifyClient(*resp); err != nil {
			log.Println("error verifying client")
			os.Exit(1)
		}
	}

	if err := fingerprint.Store(); err != nil {
		log.Println("error saving the fingerprint. Next login will require a new SMS code.", err)
	}

	log.Println("Logged in successfully")

	// TODO: Fetch all devices and prompt user to select one

	// TODO: Begin live view for selected device
}

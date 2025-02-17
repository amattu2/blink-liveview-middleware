package account

import (
	"blink-liveview-websocket/common"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// Fetches the list of Blink devices and prints them to the console
// Select one of the devices to start a liveview stream
//
// token: the Blink API token
//
// accountId: the Blink account ID
//
// region: the Blink API region
func Run(token string, accountId int, region string) {
	baseUrl := common.GetApiUrl(region)
	homescreenUrl := fmt.Sprintf("%s/api/v4/accounts/%d/homescreen", baseUrl, accountId)
	devices, err := common.Homescreen(homescreenUrl, token)
	if err != nil {
		log.Println("error getting homescreen", err)
		os.Exit(1)
	}

	fmt.Println("Select a device to start a liveview stream:")
	output, options := common.PrintDeviceOptions(devices)
	if len(options) == 0 {
		log.Println("no devices found")
		os.Exit(1)
	} else {
		fmt.Println(output)
	}

getDevice:
	fmt.Print("Device number: ")
	var deviceNumber int
	if _, err = fmt.Scanln(&deviceNumber); err != nil {
		log.Println("error reading device number", err)
		os.Exit(1)
	}
	fmt.Println()

	if deviceNumber < 1 || deviceNumber > len(options) {
		log.Println("invalid device number")
		goto getDevice
	}

	device := options[deviceNumber-1]
	log.Printf("Selected device: %s\n", device.FormattedName)

	ffplayCmd := exec.Command("ffplay",
		"-f", "mpegts",
		"-err_detect", "ignore_err",
		"-window_title", "Blink Liveview Middleware",
		"-",
	)
	inputPipe, err := ffplayCmd.StdinPipe()
	if err != nil {
		log.Println("error creating ffplay stdin pipe", err)
	}

	if err := ffplayCmd.Start(); err != nil {
		log.Println("error starting ffplay", err)
	}
	defer ffplayCmd.Process.Kill()

	ctx, cancelCtx := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Received SIGINT")
		cancelCtx()
	}()

	common.Livestream(ctx, common.AccountDetails{
		Region:     region,
		Token:      token,
		DeviceType: device.DeviceType,
		AccountId:  accountId,
		NetworkId:  device.NetworkId,
		CameraId:   device.DeviceId,
	}, inputPipe)

	inputPipe.Close()
	if err := ffplayCmd.Wait(); err != nil {
		log.Println("error waiting for ffplay", err)
	}
}

// Authenticates with the Blink API using the provided email and password
// and fetches the list of Blink devices
// Select one of the devices to start a liveview stream
//
// email: the Blink account email address
//
// password: the Blink account password
func RunWithCredentials(email string, password string) {
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
			log.Println("error verifying pin", err)
			os.Exit(1)
		}
	}

	if err := fingerprint.Store(); err != nil {
		log.Println("error saving the fingerprint. Next login will require a new SMS code.", err)
	}

	log.Printf("Logged in successfully. Token: %s, AccountId: %d, Region: %s\n", resp.Auth.Token, resp.Account.AccountId, resp.Account.Tier)

	Run(resp.Auth.Token, resp.Account.AccountId, resp.Account.Tier)
}

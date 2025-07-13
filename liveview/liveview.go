package liveview

import (
	"blink-liveview-websocket/common"
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
)

func Run(region string, token string, deviceType string, accountId int, networkId int, cameraId int) {
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

	accountDetails := common.AccountDetails{
		Region:     region,
		Token:      token,
		DeviceType: deviceType,
		AccountId:  accountId,
		NetworkId:  networkId,
		CameraId:   cameraId,
	}
	if err := common.Livestream(ctx, accountDetails, inputPipe); err != nil {
		log.Println("error during livestream", err)
	}

	inputPipe.Close()
	if err := ffplayCmd.Wait(); err != nil {
		log.Println("error waiting for ffplay", err)
	}
}

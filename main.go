package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"mlussi90/go-mailshield/config"
	imaputil "mlussi90/go-mailshield/imap"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		fmt.Printf("error loading config: %v\n", err)
		return
	}
	fmt.Println("config loaded")

	pollInterval, _ := time.ParseDuration(cfg.PollInterval)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for _, acc := range cfg.Accounts {
		go imaputil.ProcessAccount(ctx, acc, pollInterval)
	}

	<-ctx.Done()
	fmt.Println("shutdown")
}

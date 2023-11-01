package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/ostcar/calendar/model"
	"github.com/ostcar/calendar/web"
)

const serverAddr = ":8090"

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := interruptContext()
	defer cancel()

	// TODO: Maybe load location from request and let the model work with utc?
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return fmt.Errorf("loading timezone: %w", err)
	}

	model := model.New(location)

	if err := web.Run(ctx, serverAddr, model); err != nil {
		return fmt.Errorf("running http server: %w", err)
	}
	return nil
}

// interruptContext works like signal.NotifyContext
//
// In only listens on os.Interrupt. If the signal is received two times,
// os.Exit(1) is called.
func interruptContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		cancel()

		// If the signal was send for the second time, make a hard cut.
		<-sigint
		os.Exit(1)
	}()
	return ctx, cancel
}

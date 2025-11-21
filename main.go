package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"planner/internal/candles"
	"planner/internal/watcher"
	"syscall"
	"time"
)

func main() {
	w := watcher.New("HYPE")

	if err := w.StartRatioCalculator(); err != nil {
		log.Fatalf("Failed to start watcher: %v", err)
	}
	defer w.Stop()

	builder := candles.NewBuilder(1*time.Minute, w.RatioChannel())
	builder.Start()
	defer builder.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Started, waiting for candles...")
	for {
		select {
		case candle := <-builder.CandleChannel():
			prettyJSON, _ := json.MarshalIndent(candle, "", "  ")
			fmt.Printf("Candle:\n%s\n\n", string(prettyJSON))
		case <-sigChan:
			fmt.Println("\nShutting down...")
			return
		}
	}
}

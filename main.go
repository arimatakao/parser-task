package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/arimatakao/parser-task/parser"
)

func gracefulShutdown(p *parser.Parser, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	stop() // Allow Ctrl+C to force shutdown

	fmt.Println("shutdown triggered")

	// The context is used to inform the parser it has 3 seconds to finish
	// the parsing process it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		fmt.Println("shutdown parser with error: %w ", err)
	}

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	fileNames := map[string]string{"file1.txt": "out1.txt", "file2.txt": "out2.txt", "file3.txt": "out3.txt"}

	app := parser.New()

	done := make(chan bool, 1)
	isEndParsing := make(chan bool, 1)

	go gracefulShutdown(app, done)

	err := app.Run(fileNames, isEndParsing)
	if err != nil {
		fmt.Println("error while parsing:", err)
		return
	}

	select {
	case <-isEndParsing:
		fmt.Println("parsing finished successfully")
	case <-done:
		fmt.Println("shutdown signal received during parsing")
	}
}

package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/arimatakao/parser-task/parser"
)

func gracefulShutdown(ctx context.Context, p *parser.Parser) {
	// Listen for the interrupt signal.
	<-ctx.Done()

	fmt.Println("shutdown triggered")

	// The context is used to inform the parser it has 4 seconds to finish
	// the parsing process it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		fmt.Println("shutdown parser with error: %w ", err)
	}
}

func main() {
	fileNames := map[string]string{"file1.txt": "out1.txt", "file2.txt": "out2.txt", "file3.txt": "out3.txt"}

	app := parser.New()

	isEndParsing := make(chan bool, 1)
	defer close(isEndParsing)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go gracefulShutdown(ctx, app)

	err := app.Run(ctx, fileNames, isEndParsing)
	if err != nil {
		fmt.Println("error while parsing:", err)
		return
	}

	<-isEndParsing
	fmt.Println("parsing finished successfully")
}

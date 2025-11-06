package printer

import (
	"context"
	"fmt"
)

type Printerer interface {
	Print(s <-chan string) <-chan string
	Close(ctx context.Context) error
}

type Printer struct {
}

func New() *Printer {
	return &Printer{}
}

func (p *Printer) Close(ctx context.Context) error {
	return nil
}

func (p *Printer) Print(s <-chan string) <-chan string {
	copyStr := make(chan string)

	go func() {
		for {
			value, ok := <-s
			if !ok {
				close(copyStr)
				return
			}
			fmt.Println(value)
			copyStr <- value
		}
	}()

	return copyStr
}

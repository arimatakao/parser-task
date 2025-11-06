package processor

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"
)

type Message struct {
	Msg       string     `json:"message"`
	Timestamp *time.Time `json:"timestamp"`
}

type Processorer interface {
	ProcessWithDelay(ctx context.Context, source <-chan string) (<-chan string, <-chan error)
	Close(ctx context.Context) error
}

type Processor struct {
}

func New() *Processor {
	return &Processor{}
}

func (p *Processor) Close(ctx context.Context) error {
	return nil
}

func (p *Processor) ProcessWithDelay(ctx context.Context, source <-chan string) (<-chan string, <-chan error) {
	parsedMsgChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(parsedMsgChan)
		defer close(errChan)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-source:
				if !ok {
					return
				}

				message := &Message{}
				err := json.Unmarshal([]byte(line), message)
				if err != nil {
					errChan <- err
					return
				}
				time.Sleep(time.Duration(rand.Intn(5) * int(time.Second)))
				parsedMsgChan <- "[" + message.Timestamp.Format(time.DateTime) + "]: " + message.Msg
			}
		}
	}()
	return parsedMsgChan, errChan
}

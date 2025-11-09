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

type ProcessorFactoryer interface {
	Create(ctx context.Context) *Processor
	Close(ctx context.Context) error
}

type ProcessorFactory struct {
	processors []*Processor
}

func New() *ProcessorFactory {
	return &ProcessorFactory{}
}

func (f *ProcessorFactory) Close(ctx context.Context) error {
	for _, processor := range f.processors {
		err := processor.Close(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *ProcessorFactory) Create(ctx context.Context) *Processor {
	newProcessor := &Processor{
		ctx:           ctx,
		parsedMsgChan: make(chan string),
		errChan:       make(chan error),
	}

	f.processors = append(f.processors, newProcessor)
	return newProcessor
}

type Processor struct {
	ctx           context.Context
	parsedMsgChan chan string
	errChan       chan error
	isClose       bool
}

func (p *Processor) Close(ctx context.Context) error {
	if !p.isClose {
		close(p.parsedMsgChan)
		close(p.errChan)
		p.isClose = true
	}
	return nil
}

func (p *Processor) ProcessWithDelay(ctx context.Context, source <-chan string) (<-chan string, <-chan error) {

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-source:
				if !ok {
					p.isClose = true
					close(p.parsedMsgChan)
					close(p.errChan)
					return
				}

				message := &Message{}
				err := json.Unmarshal([]byte(line), message)
				if err != nil {
					p.errChan <- err
					return
				}
				time.Sleep(time.Duration(rand.Intn(5) * int(time.Second)))
				p.parsedMsgChan <- "[" + message.Timestamp.Format(time.DateTime) + "]: " + message.Msg
			}
		}
	}()
	return p.parsedMsgChan, p.errChan
}

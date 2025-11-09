package printer

import (
	"context"
	"fmt"
)

type Printerer interface {
	Print(s <-chan string) <-chan string
	Close(ctx context.Context) error
}

type PrinterFactoryer interface {
	Create(ctx context.Context) *Printer
	Close(ctx context.Context) error
}

type Printer struct {
	ctx     context.Context
	copyStr chan string
	isClose bool
}

type PrinterFactory struct {
	printers []*Printer
}

func New() *PrinterFactory {
	return &PrinterFactory{}
}

func (f *PrinterFactory) Create(ctx context.Context) *Printer {
	newPrinter := &Printer{
		ctx:     ctx,
		copyStr: make(chan string),
	}

	f.printers = append(f.printers, newPrinter)
	return newPrinter
}

func (f *PrinterFactory) Close(ctx context.Context) error {
	for _, printer := range f.printers {
		err := printer.Close(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Printer) Close(ctx context.Context) error {
	if !p.isClose {
		close(p.copyStr)
		p.isClose = true
	}
	return nil
}

func (p *Printer) Print(s <-chan string) <-chan string {
	go func() {
		for {
			value, ok := <-s
			if !ok {
				p.isClose = true
				close(p.copyStr)
				return
			}
			fmt.Println(value)
			p.copyStr <- value
		}
	}()

	return p.copyStr
}

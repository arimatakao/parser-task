package parser

import (
	"context"
	"sync"
	"time"

	"github.com/arimatakao/parser-task/filereader"
	"github.com/arimatakao/parser-task/filewriter"
	"github.com/arimatakao/parser-task/printer"
	"github.com/arimatakao/parser-task/processor"
)

type Message struct {
	Message   string
	Timestamp time.Time
}

type Parser struct {
	fr      filereader.FileReaderFactoryer
	process processor.ProcessorFactoryer
	print   printer.PrinterFactoryer
	fw      filewriter.FileWriterFactoryer
}

func New() *Parser {
	return &Parser{
		fr:      filereader.New(),
		process: processor.New(),
		print:   printer.New(),
		fw:      filewriter.New(),
	}
}

func (p *Parser) Shutdown(ctx context.Context) error {
	err := p.fr.Close(ctx)
	if err != nil {
		return err
	}
	err = p.fw.Close(ctx)
	if err != nil {
		return err
	}

	err = p.print.Close(ctx)
	if err != nil {
		return err
	}

	err = p.process.Close(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (p *Parser) Run(ctx context.Context, fileNames map[string]string, isDone chan bool) error {
	defer func() { isDone <- true }()

	var errRun error

	var wg sync.WaitGroup
	wg.Add(len(fileNames))
	for inFileName, outFileName := range fileNames {

		inChan, errRead := p.fr.Create(ctx).ReadToChan(inFileName)
		parsedJsonChan, errParsing := p.process.Create(ctx).ProcessWithDelay(ctx, inChan)
		outContent := p.print.Create(ctx).Print(parsedJsonChan)
		toFile := make(chan string)
		writingStatus, errSaving := p.fw.Create(ctx).WriteToFile(ctx, toFile, outFileName)

		go func() {
			isClosed := false
			for {
				select {
				case contentOut, more := <-outContent:
					if more {
						toFile <- contentOut
					} else if !isClosed {
						isClosed = true
						close(toFile)
					}

				case isFileSaved, more := <-writingStatus:
					if !more {
						wg.Done()
						return
					}
					if isFileSaved {
						wg.Done()
						return
					}
				case err := <-errRead:
					if err != nil {
						errRun = err
						wg.Done()
						return
					}
				case err := <-errParsing:
					if err != nil {
						errRun = err
						wg.Done()
						return
					}
				case err := <-errSaving:
					if err != nil {
						errRun = err
						wg.Done()
						return
					}
				}
			}
		}()
	}
	wg.Wait()

	return errRun
}

package filewriter

import (
	"bufio"
	"context"
	"os"
)

type FileWriterer interface {
	WriteToFile(ctx context.Context, contentChan <-chan string, fileName string) (<-chan bool, <-chan error)
	Close(ctx context.Context) error
}

type FileWriterFactoryer interface {
	Create(ctx context.Context) *FileWriter
	Close(ctx context.Context) error
}

type FileWriterFactory struct {
	fw []*FileWriter
}

func New() *FileWriterFactory {
	return &FileWriterFactory{}
}

func (f *FileWriterFactory) Create(ctx context.Context) *FileWriter {
	newWriter := &FileWriter{
		ctx:        ctx,
		statusChan: make(chan bool),
		errChan:    make(chan error),
	}
	return newWriter
}

func (f *FileWriterFactory) Close(ctx context.Context) error {
	for _, writer := range f.fw {
		err := writer.Close(ctx)
		if err != nil {
			return err
		}

	}
	return nil
}

type FileWriter struct {
	ctx        context.Context
	statusChan chan bool
	errChan    chan error
	isClose    bool
}

func (fw *FileWriter) WriteToFile(ctx context.Context, contentChan <-chan string, fileName string) (<-chan bool, <-chan error) {

	go func() {
		file, err := os.Create(fileName)
		if err != nil {
			fw.errChan <- err
		}
		defer file.Close()
		defer close(fw.statusChan)
		defer close(fw.errChan)

		w := bufio.NewWriter(file)
		for {
			select {
			case <-ctx.Done():
				w.Flush()
				return
			case chunk, ok := <-contentChan:
				if !ok {
					w.Flush()
					fw.isClose = true
					fw.statusChan <- true
					return
				}
				_, err := w.WriteString(chunk + "\n")
				if err != nil {
					fw.errChan <- err
				}
			}
		}

	}()

	return fw.statusChan, fw.errChan
}

func (fw *FileWriter) Close(ctx context.Context) error {

	return nil
}

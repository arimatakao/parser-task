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

type FileWriter struct {
}

func New() *FileWriter {
	return &FileWriter{}
}

func (fw *FileWriter) WriteToFile(ctx context.Context, contentChan <-chan string, fileName string) (<-chan bool, <-chan error) {
	statusChan := make(chan bool)
	errChan := make(chan error)

	go func() {
		file, err := os.Create(fileName)
		if err != nil {
			errChan <- err
		}
		defer file.Close()
		defer close(statusChan)
		defer close(errChan)

		w := bufio.NewWriter(file)
		for {
			select {
			case <-ctx.Done():
				w.Flush()
				return
			case chunk, ok := <-contentChan:
				if !ok {
					w.Flush()
					statusChan <- true
					return
				}
				_, err := w.WriteString(chunk + "\n")
				if err != nil {
					errChan <- err
				}
			}
		}

	}()

	return statusChan, errChan
}

func (fw *FileWriter) Close(ctx context.Context) error {
	return nil
}

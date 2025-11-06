package filereader

import (
	"bufio"
	"context"
	"io"
	"os"
)

type FileReaderer interface {
	ReadToChan(ctx context.Context, fileName string) (<-chan string, <-chan error)
	Close(ctx context.Context) error
}

type FileReader struct {
}

func New() *FileReader {
	return &FileReader{}
}

func (fr *FileReader) ReadToChan(ctx context.Context, fileName string) (<-chan string, <-chan error) {
	errChan := make(chan error)
	resChan := make(chan string)

	go func() {
		file, err := os.Open(fileName)
		if err != nil {
			errChan <- err
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		for {
			select {
			case <-ctx.Done(): // if cancel() execute
				return
			default:
				line, err := reader.ReadString('\n')

				if err == io.EOF {
					close(resChan)
					return
				} else if err != nil {
					errChan <- err
					continue
				}
				resChan <- line
			}

		}
	}()

	return resChan, errChan
}

func (fr *FileReader) Close(ctx context.Context) error {
	return nil
}

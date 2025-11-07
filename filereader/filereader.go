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

type FileReaderFactoryer interface {
	Create(ctx context.Context) *FileReader
	Close(ctx context.Context) error
}

type FileReaderFactory struct {
	readers []*FileReader
}

func New() *FileReaderFactory {
	return &FileReaderFactory{}
}

func (f *FileReaderFactory) Close(ctx context.Context) error {
	for _, reader := range f.readers {
		err := reader.Close(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FileReaderFactory) Create(ctx context.Context) *FileReader {
	newReader := &FileReader{
		ctx:     ctx,
		errChan: make(chan error),
		resChan: make(chan string),
		isClose: false,
	}

	f.readers = append(f.readers, newReader)
	return newReader
}

type FileReader struct {
	ctx     context.Context
	errChan chan error
	resChan chan string
	isClose bool
}

func (fr *FileReader) ReadToChan(fileName string) (<-chan string, <-chan error) {

	go func() {
		file, err := os.Open(fileName)
		if err != nil {
			fr.errChan <- err
			return
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		for {
			select {
			case <-fr.ctx.Done(): // if cancel() execute
				return
			default:
				line, err := reader.ReadString('\n')

				if err == io.EOF {
					fr.isClose = true
					close(fr.resChan)
					return
				} else if err != nil {
					fr.errChan <- err
					return
				}
				fr.resChan <- line
			}

		}
	}()

	return fr.resChan, fr.errChan
}

func (fr *FileReader) Close(ctx context.Context) error {
	if !fr.isClose {
		fr.isClose = true
		close(fr.resChan)
		close(fr.errChan)
	}
	return nil
}

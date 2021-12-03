package storage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

type URLStoragePersister interface {
	Store(id string, url string) error
	Load(dest map[string]string) error
}

type plainTextFileURLStoragePersister struct {
	mx       sync.Mutex
	filename string
}

func createNewPlainTextFileURLStoragePersister(filename string) *plainTextFileURLStoragePersister {
	return &plainTextFileURLStoragePersister{
		filename: filename,
	}
}

func (p *plainTextFileURLStoragePersister) Store(id string, url string) error {
	p.mx.Lock()
	defer p.mx.Unlock()
	file, err := os.OpenFile(p.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	_, err = fmt.Fprintf(w, "%s\t%s\n", id, url)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (p *plainTextFileURLStoragePersister) Load(dest map[string]string) error {
	p.mx.Lock()
	defer p.mx.Unlock()
	file, err := os.Open(p.filename)
	// файла нет, выходим
	if err != nil {
		return nil
	}
	defer file.Close()
	s := bufio.NewScanner(file)

	for s.Scan() {
		dataStr := s.Text()
		splittedData := strings.Split(dataStr, "\t")
		if len(splittedData) != 2 {
			return errors.New("invalid string in url storage file")
		}
		dest[splittedData[0]] = splittedData[1]
	}

	if err := s.Err(); err != nil {
		return err
	}

	return nil
}

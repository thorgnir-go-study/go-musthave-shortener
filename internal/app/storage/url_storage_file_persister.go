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
	Store(entity URLEntity) error
	Load(dest map[string]URLEntity) error
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

func (p *plainTextFileURLStoragePersister) Store(entity URLEntity) error {
	// тут возможны разные подходы, в зависимости от предполагаемой нагрузки
	// если предположить, что запись будет частой, то имеет смысл держать файл открытым и в структуру добавить writer
	// текущая реализация для варианта "пишем редко"
	p.mx.Lock()
	defer p.mx.Unlock()
	file, err := os.OpenFile(p.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	_, err = fmt.Fprintf(w, "%s\t%s\t%s\n", entity.ID, entity.UserID, entity.OriginalURL)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (p *plainTextFileURLStoragePersister) Load(dest map[string]URLEntity) error {
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
		if len(splittedData) != 3 {
			return errors.New("invalid string in url storage file")
		}
		dest[splittedData[0]] = URLEntity{
			ID:          splittedData[0],
			OriginalURL: splittedData[2],
			UserID:      splittedData[1],
		}
	}

	if err := s.Err(); err != nil {
		return err
	}

	return nil
}

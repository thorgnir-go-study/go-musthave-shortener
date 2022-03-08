package repository

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type inMemoryRepoFilePersister interface {
	Store(entity URLEntity) error
	Load(dest map[string]URLEntity) error
}

type inMemoryRepoFilePersisterPlain struct {
	mx       sync.Mutex
	filename string
}

func createNewInMemoryRepoFilePersisterPlain(filename string) *inMemoryRepoFilePersisterPlain {
	return &inMemoryRepoFilePersisterPlain{
		filename: filename,
	}
}

func (p *inMemoryRepoFilePersisterPlain) Store(entity URLEntity) error {
	// тут возможны разные подходы, в зависимости от предполагаемой нагрузки
	// если предположить, что запись будет частой, то имеет смысл держать файл открытым и в структуру добавить writer
	// текущая реализация для варианта "пишем редко"
	p.mx.Lock()
	defer p.mx.Unlock()
	file, err := os.OpenFile(p.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%t\n", entity.ID, entity.UserID, entity.OriginalURL, entity.Deleted)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (p *inMemoryRepoFilePersisterPlain) Load(dest map[string]URLEntity) error {
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
		if len(splittedData) != 4 {
			return errors.New("invalid string in url repository file")
		}

		isDeleted, err := strconv.ParseBool(splittedData[3])
		if err != nil {
			return fmt.Errorf("error while parsing deleted flag; %w", err)
		}
		dest[splittedData[0]] = URLEntity{
			ID:          splittedData[0],
			OriginalURL: splittedData[2],
			UserID:      splittedData[1],
			Deleted:     isDeleted,
		}
	}

	if err = s.Err(); err != nil {
		return err
	}

	return nil
}

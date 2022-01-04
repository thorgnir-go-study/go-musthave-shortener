package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/shortener"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"log"
)

var (
	serverAddressFlag *string
	baseURLFlag       *string
	storagePathFlag   *string
	databaseDsnFlag   *string
)

func init() {
	serverAddressFlag = flag.String("a", "", "Server address. If not set in CLI or env variable SERVER_ADDRESS defaults to ':8080'")
	baseURLFlag = flag.String("b", "", "Base URL. If not set in CLI or env variable BASE_URL defaults to http://localhost:8080")
	storagePathFlag = flag.String("f", "", "File storage path. If not set in CLI or env variable FILE_STORAGE_PATH storage will be non-persistent")
	databaseDsnFlag = flag.String("d", "", "Database DSN. If not set in CLI or env variable DATABASE_DSN db is not used")
}

func createRepository(cfg config.Config) (storage.URLStorager, error) {
	if cfg.DatabaseDSN != "" {
		return createDBStorage(cfg)
	}
	return createInMemoryStorage(cfg)
}

func createInMemoryStorage(cfg config.Config) (storage.URLStorager, error) {
	var options []storage.MapURLStorageOption
	if cfg.StorageFilePath != "" {
		options = append(options, storage.WithFilePersistance(cfg.StorageFilePath))
	}

	urlStorage, err := storage.NewMapURLStorage(options...)
	if err != nil {
		return nil, err
	}
	return urlStorage, nil
}

func createDBStorage(cfg config.Config) (storage.URLStorager, error) {
	urlStorage, err := storage.NewDBURLStorage(cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	return urlStorage, nil
}

func main() {
	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalln("Error parsing config")
	}

	flag.Parse()
	if *serverAddressFlag != "" {
		cfg.ServerAddress = *serverAddressFlag
	}

	if *baseURLFlag != "" {
		cfg.BaseURL = *baseURLFlag
	}

	if *storagePathFlag != "" {
		cfg.StorageFilePath = *storagePathFlag
	}

	if *databaseDsnFlag != "" {
		cfg.DatabaseDSN = *databaseDsnFlag
	}

	urlStorage, err := createRepository(cfg)
	if err != nil {
		log.Fatal(err)
	}

	idGenerator := shortener.NewRandomStringURLIDGenerator(10)

	app.StartURLShortenerServer(cfg, urlStorage, idGenerator)
}

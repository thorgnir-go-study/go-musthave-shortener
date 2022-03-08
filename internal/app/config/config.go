package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress            string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL                  string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	StorageFilePath          string `env:"FILE_STORAGE_PATH"`
	AuthSecretKey            string `env:"AUTH_SECRET_KEY" envDefault:"very very secret key"`
	DatabaseDSN              string `env:"DATABASE_DSN"`
	ShortenBatchSize         int    `env:"SHORTEN_BATCH_SIZE" envDefault:"100"`
	ShortURLIdentifierLength int    `env:"URL_ID_LENGTH" envDefault:"10"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("configuration failure: failed to parse environment. %w", err)
	}

	// объявляем флаги, дефолтными значениями указываем то, что уже в конфиге (заполнено из env-переменных). Таким образом:
	// - если не передан ни флаг, ни установлена переменная окружения - используется envDefault заданный в структурных тегах
	// - если передана переменная окружения, но не передан флаг - будет использоваться значение переменной окружения
	// - если передан флаг - он оверрайдит и дефолты и значения переменных окружения
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Server address. If not set in CLI or env variable SERVER_ADDRESS defaults to ':8080'")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL. If not set in CLI or env variable BASE_URL defaults to http://localhost:8080")
	flag.StringVar(&cfg.StorageFilePath, "f", cfg.StorageFilePath, "File repository path. If not set in CLI or env variable FILE_STORAGE_PATH repository will be non-persistent")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN. If not set in CLI or env variable DATABASE_DSN db is not used")
	flag.IntVar(&cfg.ShortenBatchSize, "shorten-batch-size", cfg.ShortenBatchSize, "Batch size for shorten. If not set in CLI or env variable SHORTEN_BATCH_SIZE defaults to 100")
	flag.IntVar(&cfg.ShortURLIdentifierLength, "url-id-length", cfg.ShortURLIdentifierLength, "Short url id length. If not set in CLI or env variable URL_ID_LENGTH defaults to 10")

	flag.Parse()

	return cfg, nil

}

package config

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	StorageFilePath string `env:"FILE_STORAGE_PATH"`
	AuthSecretKey   string `env:"AUTH_SECRET_KEY" envDefault:"very very secret key"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

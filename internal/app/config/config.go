package config

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	StorageFilePath string `env:"FILE_STORAGE_PATH" envDefault:"url_storage.txt"`
}

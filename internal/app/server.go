package app

import (
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/handlers"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"log"
	"net/http"
)

//StartURLShortenerServer старт нового сервера сокращения ссылок
func StartURLShortenerServer(domain string, storage storage.URLStorage) {
	r := handlers.NewRouter(storage)
	log.Fatal(http.ListenAndServe(domain, r))
}

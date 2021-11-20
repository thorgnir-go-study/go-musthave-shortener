package main

import (
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
)

func main() {
	urlStorage := storage.CreateMapURLStorage()
	app.StartURLShortenerServer(8080, urlStorage)
}

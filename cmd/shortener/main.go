package main

import "github.com/thorgnir-go-study/go-musthave-shortener/internal/app"

func main() {
	storage := app.CreateMapURLStorage()
	app.StartURLShortenerServer(8080, storage)
}

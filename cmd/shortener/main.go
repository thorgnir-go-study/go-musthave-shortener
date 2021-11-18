package main

import "github.com/thorgnir-go-study/go-musthave-shortener/internal/app"

func main() {
	storage := app.CreateMapUrlStorage()
	app.StartUrlShortenerServer(8080, storage)
}

package app

import (
	"fmt"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/handlers"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"net/http"
)

func StartURLShortenerServer(port uint16, storage storage.URLStorage) {
	http.HandleFunc("/", handlers.RootHanlder(storage))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

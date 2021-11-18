package app

import (
	"fmt"
	"net/http"
)

func registerHandlers() {
	http.HandleFunc("/", rootHandler)
}

func StartUrlShortenerServer(port uint16, storage UrlStorage) {
	urlStorage = storage
	registerHandlers()
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

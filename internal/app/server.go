package app

import (
	"fmt"
	"net/http"
)

func registerHandlers() {
	http.HandleFunc("/", rootHandler)
}

func StartURLShortenerServer(port uint16, storage URLStorage) {
	urlStorage = storage
	registerHandlers()
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

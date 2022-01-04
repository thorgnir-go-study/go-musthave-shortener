package random

import (
	"math/rand"
	"time"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func String(n int) string {
	// хороший пост про возможные оптимизации производительности генерации строки тут: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[seededRand.Int63()%int64(len(letters))]
	}
	return string(b)
}

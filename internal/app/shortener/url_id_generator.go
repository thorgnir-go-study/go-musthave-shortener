package shortener

import "math/rand"

type URLIDGenerator interface {
	// GenerateURLID генерирует идентификатор сокращенной ссылки
	GenerateURLID(originalURL string) string
}

type RandomStringURLIDGenerator struct {
	length int
}

func NewRandomStringURLIDGenerator(length int) *RandomStringURLIDGenerator {
	return &RandomStringURLIDGenerator{length: length}
}

func (g *RandomStringURLIDGenerator) GenerateURLID(originalURL string) string {
	return randomString(g.length)
}

var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func randomString(n int) string {
	// хороший пост про возможные оптимизации производительности генерации строки тут: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}

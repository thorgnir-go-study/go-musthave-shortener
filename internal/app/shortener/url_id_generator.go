package shortener

import "github.com/thorgnir-go-study/go-musthave-shortener/internal/app/random"

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
	return random.String(g.length)
}

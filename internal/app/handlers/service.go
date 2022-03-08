package handlers

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/shortener"
	"time"
)

type deleteURLsRequest struct {
	ids    []string
	userID string
}

type Service struct {
	Repository      repository.URLRepository
	IDGenerator     shortener.URLIDGenerator
	Config          config.Config
	deleteURLsReqCh chan<- deleteURLsRequest
}

func NewService(repository repository.URLRepository, IDGenerator shortener.URLIDGenerator, config config.Config) *Service {

	s := &Service{Repository: repository, IDGenerator: IDGenerator, Config: config}
	s.deleteURLsReqCh = s.startDeleteURLsWorkers(context.Background(), 5)

	return s
}

func (s *Service) startDeleteURLsWorkers(ctx context.Context, count int) chan<- deleteURLsRequest {
	ch := make(chan deleteURLsRequest, count*2)
	for i := 0; i < count; i++ {
		workerID := fmt.Sprintf("DeleteURLsWorker#%d", i+1)
		go func() {
			log.Info().Str("worker", workerID).Msg("starting delete urls worker")
			for {
				select {
				case <-ctx.Done():
					log.Info().Str("worker", workerID).Msg("stopping delete urls worker")
				case req := <-ch:
					go func() {
						innerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
						defer cancel()

						err := s.Repository.DeleteURLs(innerCtx, req.userID, req.ids)
						if err != nil {
							log.Error().
								Err(err).
								Str("worker", workerID).
								Strs("ids", req.ids).
								Str("userID", req.userID).
								Msg("error while deleting user urls")
							return
						}
						log.Info().
							Str("worker", workerID).
							Strs("ids", req.ids).
							Str("userID", req.userID).
							Msg("urls deleted")
					}()

				}
			}
		}()
	}
	return ch
}

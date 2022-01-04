package storage

import "context"

type BatchWriter interface {
	Add(ctx context.Context, e URLEntity) error
	Flush(ctx context.Context) error
}

type BatchService struct {
	batchSize  int
	buffer     []URLEntity
	repository URLStorager
}

func NewBatchService(batchSize int, repository URLStorager) *BatchService {
	return &BatchService{
		batchSize:  batchSize,
		buffer:     make([]URLEntity, 0, batchSize),
		repository: repository,
	}
}

func (s *BatchService) Add(ctx context.Context, e URLEntity) error {
	s.buffer = append(s.buffer, e)
	if cap(s.buffer) == len(s.buffer) {
		if err := s.Flush(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *BatchService) Flush(ctx context.Context) error {
	if len(s.buffer) == 0 {
		return nil
	}
	err := s.repository.StoreBatch(ctx, s.buffer)
	if err != nil {
		return err
	}
	s.buffer = s.buffer[:0]
	return nil
}

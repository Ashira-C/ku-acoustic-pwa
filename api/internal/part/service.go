package part

import "context"

// repository is defined here, by the consumer, so tests can pass a fake.
type repository interface {
	List(ctx context.Context) ([]Part, error)
}

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Part, error) {
	return s.repo.List(ctx)
}
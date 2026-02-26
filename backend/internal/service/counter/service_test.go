package counter

import "github.com/AbelHaro/url-shortener/backend/internal/repository/counter"

func provideService() (*Service, error) {
	repo := counter.NewMockRepository()
	return NewService(repo)
}

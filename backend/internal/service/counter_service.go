package service

import (
	"sync"
	"sync/atomic"

	"github.com/AbelHaro/url-shortener/backend/internal/repository"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type CounterService struct {
	mu      sync.RWMutex
	counter int64
	repo    repository.HashCounterRepository
}

func NewCounterService(repo repository.HashCounterRepository) (*CounterService, error) {
	svc := &CounterService{
		repo: repo,
	}

	hashCounter, err := repo.GetCounter()
	if err != nil {
		return nil, err
	}
	if hashCounter != nil {
		atomic.StoreInt64(&svc.counter, hashCounter.Counter)
	}

	return svc, nil
}

func (svc *CounterService) NextBase62() (string, error) {
	newVal := atomic.AddInt64(&svc.counter, 1)

	if err := svc.repo.UpdateCounter(newVal); err != nil {
		//Rollback of the updated number
		atomic.AddInt64(&svc.counter, -1)
		return "", err
	}

	return svc.ToBase62(newVal), nil
}

func (svc *CounterService) ToBase62(n int64) string {
	if n == 0 {
		return string(base62Chars[0])
	}

	var result []byte
	length := len(base62Chars)

	for n > 0 {
		result = append([]byte{base62Chars[n%int64(length)]}, result...)
		n /= int64(length)
	}

	return string(result)
}

func (svc *CounterService) PadBase62(s string, length int) string {
	if len(s) >= length {
		return s
	}
	padding := make([]byte, length-len(s))
	for i := range padding {
		padding[i] = '0'
	}
	return string(padding) + s
}

func (svc *CounterService) indexOf(char rune, s string) int {
	for i, c := range s {
		if c == char {
			return i
		}
	}
	return -1
}

func (svc *CounterService) GenerateShortHash() (string, error) {
	num, err := svc.NextBase62()
	if err != nil {
		return "", err
	}
	return svc.PadBase62(num, 7), nil
}

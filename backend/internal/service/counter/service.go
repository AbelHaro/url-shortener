package counter

import (
	"encoding/binary"
	"sync"
	"sync/atomic"

	"github.com/cyrildever/feistel"

	counterRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/counter"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Service struct {
	mu       sync.RWMutex
	counter  int64
	repo     counterRepo.Repository
	cipher   *feistel.Cipher
	maxValue uint64
}

func NewService(repo counterRepo.Repository) (*Service, error) {
	svc := &Service{
		repo:     repo,
		cipher:   feistel.NewCipher("url-shortener-secret-key-2026", 12),
		maxValue: 62 * 62 * 62 * 62 * 62 * 62 * 62,
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

func (svc *Service) NextBase62() (string, error) {
	newVal := atomic.AddInt64(&svc.counter, 1)

	if err := svc.repo.UpdateCounter(newVal); err != nil {
		atomic.AddInt64(&svc.counter, -1)
		return "", err
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(newVal))

	encrypted, err := svc.cipher.Encrypt(string(buf))
	if err != nil {
		return "", err
	}

	encryptedBytes := []byte(encrypted)
	result := binary.BigEndian.Uint64(encryptedBytes[:8]) % svc.maxValue

	return svc.ToBase62(int64(result)), nil
}

func (svc *Service) ToBase62(n int64) string {
	if n == 0 {
		return string(base62Chars[0])
	}

	var result []byte
	length := int64(len(base62Chars))

	for n > 0 {
		result = append([]byte{base62Chars[n%length]}, result...)
		n /= length
	}

	return string(result)
}

func (svc *Service) PadBase62(s string, length int) string {
	if len(s) >= length {
		return s
	}
	padding := make([]byte, length-len(s))
	for i := range padding {
		padding[i] = '0'
	}
	return string(padding) + s
}

func (svc *Service) indexOf(char rune, s string) int {
	for i, c := range s {
		if c == char {
			return i
		}
	}
	return -1
}

func (svc *Service) GenerateShortHash() (string, error) {
	num, err := svc.NextBase62()
	if err != nil {
		return "", err
	}
	return svc.PadBase62(num, 7), nil
}

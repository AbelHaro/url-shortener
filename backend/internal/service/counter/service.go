package counter

import (
	"encoding/binary"
	"sync"
	"sync/atomic"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	rangeservice "github.com/AbelHaro/url-shortener/backend/internal/service/range"
	"github.com/cyrildever/feistel"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Service struct {
	mu           sync.RWMutex
	counter      int64
	rangeSvc     *rangeservice.Service
	currentRange *domain.Range
	cipher       *feistel.Cipher
	maxValue     uint64
}

func NewService(rangeSvc *rangeservice.Service) (*Service, error) {
	svc := &Service{
		rangeSvc: rangeSvc,
		cipher:   feistel.NewCipher("url-shortener-secret-key-2026", 12),
		maxValue: 62 * 62 * 62 * 62 * 62 * 62 * 62,
	}

	rangeAllocated, err := rangeSvc.AllocateRange()
	if err != nil {
		return nil, err
	}
	if rangeAllocated != nil {
		atomic.StoreInt64(&svc.counter, int64(rangeAllocated.Start+rangeAllocated.CurrentOffset))
		svc.currentRange = rangeAllocated
	}

	return svc, nil
}

func (svc *Service) NextBase62() (string, error) {

	svc.mu.Lock()
	defer svc.mu.Unlock()
	newVal := atomic.AddInt64(&svc.counter, 1)
	if newVal >= int64(svc.currentRange.Last) {
		rangeAllocated, err := svc.rangeSvc.AllocateRange()
		if err != nil {
			return "", err
		}
		if rangeAllocated != nil {
			atomic.StoreInt64(&svc.counter, int64(rangeAllocated.Start+rangeAllocated.CurrentOffset))
			svc.currentRange = rangeAllocated
			newVal = atomic.AddInt64(&svc.counter, 1)
		} else {
			return "", err
		}
	}

	svc.mu.Unlock()

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

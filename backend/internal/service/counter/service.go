package counter

import (
	"encoding/binary"
	"sync"
	"sync/atomic"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	idsRangesService "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	"github.com/cyrildever/feistel"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Service struct {
	mu       sync.RWMutex
	counter  int64
	rangeSvc *idsRangesService.Service
	IDsRange *domain.IDsRange
	cipher   *feistel.Cipher
	maxValue uint64
}

func NewService(rangeSvc *idsRangesService.Service) (*Service, error) {
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
		svc.IDsRange = rangeAllocated
	}

	return svc, nil
}

// NextBase62 generates the next unique ID in base62 encoding. It uses an atomic counter to ensure thread safety and checks if the current counter value exceeds the allocated range. If it does, it allocates a new range and updates the counter accordingly. The generated ID is then encrypted using a Feistel cipher to add an extra layer of obfuscation before being converted to base62.
func (svc *Service) NextBase62() (string, error) {

	svc.mu.Lock()
	newVal := atomic.AddInt64(&svc.counter, 1)
	// Check if the new value exceeds the allocated range. If it does, allocate a new range and update the counter accordingly. Do this in a thread-safe way to avoid race conditions and ensure that only one goroutine can allocate a new range at a time.
	if newVal >= int64(svc.IDsRange.Last) {
		rangeAllocated, err := svc.rangeSvc.AllocateRange()
		if err != nil {
			svc.mu.Unlock()
			return "", err
		}
		if rangeAllocated != nil {
			atomic.StoreInt64(&svc.counter, int64(rangeAllocated.Start+rangeAllocated.CurrentOffset))
			svc.IDsRange = rangeAllocated
			newVal = atomic.AddInt64(&svc.counter, 1)
		} else {
			svc.mu.Unlock()
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

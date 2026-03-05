package counter

import (
	"encoding/binary"
	"testing"

	"github.com/cyrildever/feistel"

	"github.com/AbelHaro/url-shortener/backend/internal/repository/counter"
)

func provideService() (*Service, error) {
	repo := counter.NewMockRepository()
	return NewService(repo)
}

func TestService_GenerateShortHash(t *testing.T) {
	svc, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	hashes := make(map[string]bool)

	for i := 0; i < 100; i++ {
		hash, err := svc.GenerateShortHash()

		if i%10 == 0 {
			t.Logf("Generated hash %d: %s", i, hash)
		}

		if err != nil {
			t.Errorf("Service.GenerateShortHash() error = %v", err)
		}
		if len(hash) != 7 {
			t.Errorf("Service.GenerateShortHash() returned hash with length %d, want 7", len(hash))
		}
		if hashes[hash] {
			t.Errorf("Service.GenerateShortHash() generated duplicate hash: %s", hash)
		}
		hashes[hash] = true
	}
}

func TestFeistel_CollisionFree(t *testing.T) {
	cipher := feistel.NewCipher("test-key", 12)
	seen := make(map[uint64]bool)
	maxValue := uint64(62 * 62 * 62 * 62 * 62 * 62 * 62)

	for i := uint64(1); i <= 10000; i++ {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, i)

		encrypted, err := cipher.Encrypt(string(buf))
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		encryptedBytes := []byte(encrypted)
		result := binary.BigEndian.Uint64(encryptedBytes[:8]) % maxValue

		if seen[result] {
			t.Errorf("Collision detected: %d maps to %d which was already used", i, result)
		}
		seen[result] = true
	}
}

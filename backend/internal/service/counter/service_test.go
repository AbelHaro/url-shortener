package counter

import (
	"encoding/binary"
	"testing"

	idsrangesRepository "github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	idsrangesService "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	"github.com/cyrildever/feistel"
)

func provideService() (*Service, error) {
	idsRangesRepository := idsrangesRepository.NewMockRepository()
	idsRangesService := idsrangesService.NewService(idsRangesRepository)
	return NewService(idsRangesService)
}

func TestService_GenerateShortHash(t *testing.T) {
	svc, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	hashes := make(map[string]bool)

	for i := range 100 {
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

func TestService_RestartUsesNextOffset(t *testing.T) {
	tests := []struct {
		name       string
		starts     int
		wantOffset uint64
	}{
		{name: "one restart", starts: 2, wantOffset: 200},
		{name: "several restarts", starts: 5, wantOffset: 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := idsrangesRepository.NewMockRepository()
			rangeService := idsrangesService.NewService(repository)
			seenCodes := make(map[string]bool, tt.starts)
			var rangeID string
			var lastService *Service

			for start := range tt.starts {
				service, err := NewService(rangeService)
				if err != nil {
					t.Fatalf("NewService() start %d error = %v", start, err)
				}
				if start == 0 {
					rangeID = service.IDsRange.ID.String()
				} else if service.IDsRange.ID.String() != rangeID {
					t.Fatalf("start %d allocated range %s, want %s", start, service.IDsRange.ID, rangeID)
				}

				shortCode, err := service.GenerateShortHash()
				if err != nil {
					t.Fatalf("GenerateShortHash() start %d error = %v", start, err)
				}
				if seenCodes[shortCode] {
					t.Errorf("start %d generated duplicate short code %q", start, shortCode)
				}
				seenCodes[shortCode] = true
				lastService = service
			}

			if lastService.IDsRange.CurrentOffset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", lastService.IDsRange.CurrentOffset, tt.wantOffset)
			}
		})
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

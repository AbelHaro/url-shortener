package url

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	idsRangesRepository "github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	counterService "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	idsRangesService "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	"github.com/google/uuid"
)

type trackingCache struct {
	entries     map[string]*domain.URL
	getErr      error
	setErr      error
	deleteErr   error
	getCalls    int
	setCalls    int
	deleteCalls int
}

func newTrackingCache() *trackingCache {
	return &trackingCache{entries: make(map[string]*domain.URL)}
}

func (cache *trackingCache) GetByShortCode(_ context.Context, shortCode string) (*domain.URL, error) {
	cache.getCalls++
	if cache.getErr != nil {
		return nil, cache.getErr
	}
	storedURL, ok := cache.entries[shortCode]
	if !ok {
		return nil, domain.ErrCacheMiss
	}
	copyOfURL := *storedURL
	return &copyOfURL, nil
}

func (cache *trackingCache) SetByShortCode(_ context.Context, shortCode string, storedURL *domain.URL) error {
	cache.setCalls++
	if cache.setErr != nil {
		return cache.setErr
	}
	copyOfURL := *storedURL
	cache.entries[shortCode] = &copyOfURL
	return nil
}

func (cache *trackingCache) DeleteByShortCode(_ context.Context, shortCode string) error {
	cache.deleteCalls++
	if cache.deleteErr != nil {
		return cache.deleteErr
	}
	delete(cache.entries, shortCode)
	return nil
}

type failingStoreRepository struct {
	url.Repository
}

func (repository failingStoreRepository) Store(*domain.URL) (*domain.URL, error) {
	return nil, errors.New("database unavailable")
}

func TestService_OwnerScopedURLAccess(t *testing.T) {
	t.Parallel()

	service, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}
	ownerID := uuid.New()
	otherUserID := uuid.New()
	storedURL, err := service.Store("https://owner.example", ownerID)
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	tests := []struct {
		name    string
		userID  uuid.UUID
		wantErr error
	}{
		{name: "owner can access", userID: ownerID},
		{name: "other user sees not found", userID: otherUserID, wantErr: domain.ErrURLNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.FindByIDForUser(context.Background(), storedURL.ID.String(), tt.userID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByIDForUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_UpdateOriginalURLForUser(t *testing.T) {
	t.Parallel()

	service, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}
	ownerID := uuid.New()
	storedURL, err := service.Store("https://before.example", ownerID)
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}
	originalShortCode := storedURL.ShortCode

	tests := []struct {
		name        string
		userID      uuid.UUID
		originalURL string
		wantErr     error
	}{
		{name: "owner updates destination", userID: ownerID, originalURL: "https://after.example/path"},
		{name: "invalid destination", userID: ownerID, originalURL: "not-a-url", wantErr: domain.ErrInvalidURL},
		{name: "different user sees not found", userID: uuid.New(), originalURL: "https://forbidden.example", wantErr: domain.ErrURLNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedURL, err := service.UpdateOriginalURLForUser(
				context.Background(),
				storedURL.ID.String(),
				tt.userID,
				tt.originalURL,
			)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UpdateOriginalURLForUser() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if updatedURL.ID != storedURL.ID {
				t.Errorf("ID = %s, want %s", updatedURL.ID, storedURL.ID)
			}
			if updatedURL.ShortCode != originalShortCode {
				t.Errorf("ShortCode = %q, want %q", updatedURL.ShortCode, originalShortCode)
			}
			if updatedURL.OriginalURL != tt.originalURL {
				t.Errorf("OriginalURL = %q, want %q", updatedURL.OriginalURL, tt.originalURL)
			}
		})
	}
}

func provideService() (*Service, error) {
	repo := url.NewMockRepository()
	return provideServiceWithDependencies(repo, nil)
}

func provideServiceWithDependencies(repo url.Repository, cache url.Cache) (*Service, error) {
	idsRangesRepository := idsRangesRepository.NewMockRepository()
	idsRangesService := idsRangesService.NewService(idsRangesRepository)
	counterService, err := counterService.NewService(idsRangesService)

	if err != nil {
		return nil, err
	}

	return NewService(repo, cache, counterService), nil
}

func TestService_URLCacheAside(t *testing.T) {
	tests := []struct {
		name          string
		arrange       func(*testing.T, url.Repository, *trackingCache) string
		wantOriginal  string
		wantCacheSets int
		wantCached    bool
	}{
		{
			name: "cache hit",
			arrange: func(t *testing.T, _ url.Repository, cache *trackingCache) string {
				t.Helper()
				cache.entries["cached"] = &domain.URL{ShortCode: "cached", OriginalURL: "https://cached.example"}
				return "cached"
			},
			wantOriginal:  "https://cached.example",
			wantCacheSets: 0,
			wantCached:    true,
		},
		{
			name: "cache miss loads database and populates cache",
			arrange: func(t *testing.T, repository url.Repository, _ *trackingCache) string {
				t.Helper()
				storedURL, err := repository.Store(&domain.URL{ShortCode: "database", OriginalURL: "https://database.example"})
				if err != nil {
					t.Fatalf("Store() error = %v", err)
				}
				return storedURL.ShortCode
			},
			wantOriginal:  "https://database.example",
			wantCacheSets: 1,
			wantCached:    true,
		},
		{
			name: "cache failure falls back to database",
			arrange: func(t *testing.T, repository url.Repository, cache *trackingCache) string {
				t.Helper()
				cache.getErr = errors.New("Valkey unavailable")
				storedURL, err := repository.Store(&domain.URL{ShortCode: "fallback", OriginalURL: "https://fallback.example"})
				if err != nil {
					t.Fatalf("Store() error = %v", err)
				}
				return storedURL.ShortCode
			},
			wantOriginal:  "https://fallback.example",
			wantCacheSets: 0,
			wantCached:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := url.NewMockRepository()
			cache := newTrackingCache()
			shortCode := tt.arrange(t, repository, cache)
			service, err := provideServiceWithDependencies(repository, cache)
			if err != nil {
				t.Fatalf("provideServiceWithDependencies() error = %v", err)
			}

			foundURL, err := service.FindByShortCode(shortCode)
			if err != nil {
				t.Fatalf("FindByShortCode() error = %v", err)
			}
			if foundURL.OriginalURL != tt.wantOriginal {
				t.Errorf("OriginalURL = %q, want %q", foundURL.OriginalURL, tt.wantOriginal)
			}
			if cache.setCalls != tt.wantCacheSets {
				t.Errorf("cache Set calls = %d, want %d", cache.setCalls, tt.wantCacheSets)
			}
			_, cached := cache.entries[shortCode]
			if cached != tt.wantCached {
				t.Errorf("cache populated = %t, want %t", cached, tt.wantCached)
			}
		})
	}
}

func TestService_StoreDatabaseFirstAndCacheFailureIsNonBlocking(t *testing.T) {
	t.Run("database failure does not populate cache", func(t *testing.T) {
		cache := newTrackingCache()
		repository := failingStoreRepository{Repository: url.NewMockRepository()}
		service, err := provideServiceWithDependencies(repository, cache)
		if err != nil {
			t.Fatalf("provideServiceWithDependencies() error = %v", err)
		}

		_, err = service.Store("https://example.com", uuid.New())
		if !errors.Is(err, domain.ErrInternal) {
			t.Fatalf("Store() error = %v, want %v", err, domain.ErrInternal)
		}
		if cache.setCalls != 0 {
			t.Errorf("cache Set calls = %d, want 0", cache.setCalls)
		}
	})

	t.Run("cache failure preserves database success", func(t *testing.T) {
		cache := newTrackingCache()
		cache.setErr = errors.New("Valkey unavailable")
		service, err := provideServiceWithDependencies(url.NewMockRepository(), cache)
		if err != nil {
			t.Fatalf("provideServiceWithDependencies() error = %v", err)
		}

		storedURL, err := service.Store("https://example.com", uuid.New())
		if err != nil {
			t.Fatalf("Store() error = %v", err)
		}
		foundURL, err := service.repo.FindByShortCode(storedURL.ShortCode)
		if err != nil {
			t.Fatalf("FindByShortCode() error = %v", err)
		}
		if foundURL.ID != storedURL.ID {
			t.Errorf("stored ID = %s, want %s", foundURL.ID, storedURL.ID)
		}
	})
}

func TestService_URLCacheMutationConsistency(t *testing.T) {
	t.Run("update refreshes destination", func(t *testing.T) {
		cache := newTrackingCache()
		service, err := provideServiceWithDependencies(url.NewMockRepository(), cache)
		if err != nil {
			t.Fatalf("provideServiceWithDependencies() error = %v", err)
		}
		ownerID := uuid.New()
		storedURL, err := service.Store("https://before.example", ownerID)
		if err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		_, err = service.UpdateOriginalURLForUser(context.Background(), storedURL.ID.String(), ownerID, "https://after.example")
		if err != nil {
			t.Fatalf("UpdateOriginalURLForUser() error = %v", err)
		}
		if got := cache.entries[storedURL.ShortCode].OriginalURL; got != "https://after.example" {
			t.Errorf("cached OriginalURL = %q, want %q", got, "https://after.example")
		}
	})

	t.Run("failed update refresh invalidates old destination", func(t *testing.T) {
		cache := newTrackingCache()
		service, err := provideServiceWithDependencies(url.NewMockRepository(), cache)
		if err != nil {
			t.Fatalf("provideServiceWithDependencies() error = %v", err)
		}
		ownerID := uuid.New()
		storedURL, err := service.Store("https://before.example", ownerID)
		if err != nil {
			t.Fatalf("Store() error = %v", err)
		}
		cache.setErr = errors.New("Valkey SET unavailable")

		_, err = service.UpdateOriginalURLForUser(context.Background(), storedURL.ID.String(), ownerID, "https://after.example")
		if err != nil {
			t.Fatalf("UpdateOriginalURLForUser() error = %v", err)
		}
		if _, cached := cache.entries[storedURL.ShortCode]; cached {
			t.Errorf("stale short code %q remains cached", storedURL.ShortCode)
		}
	})

	deleteMethods := []struct {
		name   string
		delete func(*Service, *domain.URL) error
	}{
		{name: "by ID", delete: func(service *Service, storedURL *domain.URL) error {
			return service.DeleteByID(storedURL.ID.String())
		}},
		{name: "by ID and owner", delete: func(service *Service, storedURL *domain.URL) error {
			return service.DeleteByIDForUser(context.Background(), storedURL.ID.String(), storedURL.UserID)
		}},
		{name: "by original URL", delete: func(service *Service, storedURL *domain.URL) error {
			return service.DeleteByOriginalURL(storedURL.OriginalURL)
		}},
		{name: "by short code", delete: func(service *Service, storedURL *domain.URL) error {
			return service.DeleteByShortCode(storedURL.ShortCode)
		}},
	}

	for _, tt := range deleteMethods {
		t.Run("delete "+tt.name+" invalidates cache", func(t *testing.T) {
			cache := newTrackingCache()
			service, err := provideServiceWithDependencies(url.NewMockRepository(), cache)
			if err != nil {
				t.Fatalf("provideServiceWithDependencies() error = %v", err)
			}
			storedURL, err := service.Store(fmt.Sprintf("https://%s.example", uuid.NewString()), uuid.New())
			if err != nil {
				t.Fatalf("Store() error = %v", err)
			}

			if err := tt.delete(service, storedURL); err != nil {
				t.Fatalf("delete error = %v", err)
			}
			if _, cached := cache.entries[storedURL.ShortCode]; cached {
				t.Errorf("short code %q remains cached", storedURL.ShortCode)
			}
		})
	}

	t.Run("cache invalidation failure does not undo database delete", func(t *testing.T) {
		cache := newTrackingCache()
		service, err := provideServiceWithDependencies(url.NewMockRepository(), cache)
		if err != nil {
			t.Fatalf("provideServiceWithDependencies() error = %v", err)
		}
		storedURL, err := service.Store("https://delete.example", uuid.New())
		if err != nil {
			t.Fatalf("Store() error = %v", err)
		}
		cache.deleteErr = errors.New("Valkey unavailable")

		if err := service.DeleteByShortCode(storedURL.ShortCode); err != nil {
			t.Fatalf("DeleteByShortCode() error = %v", err)
		}
		foundURL, err := service.repo.FindByShortCode(storedURL.ShortCode)
		if err != nil {
			t.Fatalf("repository FindByShortCode() error = %v", err)
		}
		if foundURL != nil {
			t.Errorf("repository still contains short code %q", storedURL.ShortCode)
		}
	})
}

func TestService_Store(t *testing.T) {
	svc, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	tests := []struct {
		name        string
		originalURL string
		wantErr     bool
	}{
		{"valid url", "https://google.com", false},
		{"invalid url", "not a url", true},
		{"empty url", "", true},
		{"repited url", "https://google.com", false},
		{"long url", "https://google.com/dadadadada/dadadadaa", false},
		{"url with query paramaters", "https://google.com/hello?query1=hello2&query2=hello3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Store(tt.originalURL, uuid.New())
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_StoreAllowsDuplicateOriginalURLs(t *testing.T) {
	t.Parallel()

	service, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}
	ownerID := uuid.New()

	first, err := service.Store("https://duplicate.example", ownerID)
	if err != nil {
		t.Fatalf("first Store() error = %v", err)
	}
	second, err := service.Store("https://duplicate.example", ownerID)
	if err != nil {
		t.Fatalf("second Store() error = %v", err)
	}

	if first.ID == second.ID {
		t.Errorf("duplicate destinations reused ID %s", first.ID)
	}
	if first.ShortCode == second.ShortCode {
		t.Errorf("duplicate destinations reused short code %q", first.ShortCode)
	}
}

func TestService_FindByShortCode(t *testing.T) {
	svc, err := provideService()

	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	tests := []struct {
		name        string
		originalURL string
		wantErr     bool
	}{
		{"valid url", "https://google.com", false},
		{"invalid url", "not a url", true},
		{"repited url", "https://google.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlInserted, err := svc.Store(tt.originalURL, uuid.New())
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Store() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			urlFound, err := svc.FindByShortCode(urlInserted.ShortCode)

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.FindByShortCode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if urlFound != urlInserted {
				t.Errorf("Service.FindByShortCode() = %v, want %v", urlFound, urlInserted)
			}
		})
	}

}

func TestService_FindByID(t *testing.T) {
	svc, err := provideService()

	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	tests := []struct {
		name        string
		originalURL string
		wantErr     bool
	}{
		{"valid url", "https://google.com", false},
		{"invalid url", "not a url", true},
		{"repited url", "https://google.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlInserted, err := svc.Store(tt.originalURL, uuid.New())

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Store() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			urlFound, err := svc.FindByID(urlInserted.ID.String())

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if urlFound != urlInserted {
				t.Errorf("Service.FindByID() = %v, want %v", urlFound, urlInserted)
			}

		})

	}
}

func TestService_FindByOriginalURL(t *testing.T) {
	svc, err := provideService()

	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	tests := []struct {
		name        string
		originalURL string
		wantErr     bool
	}{
		{"valid url", "https://google.com", false},
		{"invalid url", "not a url", true},
		{"repited url", "https://google.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlInserted, err := svc.Store(tt.originalURL, uuid.New())

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Store() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			urlFound, err := svc.FindByOriginalURL(urlInserted.OriginalURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.FindByOriginalURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if urlFound.OriginalURL != urlInserted.OriginalURL {
				t.Errorf("Service.FindByOriginalURL().OriginalURL = %q, want %q", urlFound.OriginalURL, urlInserted.OriginalURL)
			}
		})
	}
}

func TestService_FindAllByUserID(t *testing.T) {
	service, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}
	ownerID := uuid.New()
	for _, originalURL := range []string{"https://one.example", "https://two.example"} {
		if _, err := service.Store(originalURL, ownerID); err != nil {
			t.Fatalf("Store() error = %v", err)
		}
	}
	if _, err := service.Store("https://other.example", uuid.New()); err != nil {
		t.Fatalf("Store() for other owner error = %v", err)
	}

	urls, err := service.FindAllByUserID(ownerID)
	if err != nil {
		t.Fatalf("FindAllByUserID() error = %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("FindAllByUserID() returned %d URLs, want 2", len(urls))
	}
}

func TestService_DeleteByID(t *testing.T) {
	svc, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	tests := []struct {
		name        string
		originalURL string
		id          uuid.UUID
		wantErr     bool
	}{
		{"Stored url is deleted", "https://google.com", uuid.Nil, false},
		{"Not stored url", "https://google.com", uuid.New(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.id == uuid.Nil {
				urlInserted, err := svc.Store(tt.originalURL, uuid.New())
				if (err != nil) != tt.wantErr {
					t.Errorf("Service.Store() error = %v, wantErr %v", err, tt.wantErr)
				}
				tt.id = urlInserted.ID
			}
			err := svc.DeleteByID(tt.id.String())

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.DeleteById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_DeleteByOriginalURL(t *testing.T) {
	svc, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	tests := []struct {
		name        string
		originalURL string
		wantErr     bool
	}{
		{"Stored url is deleted", "https://google.com", false},
		{"Not stored url", "https://notfound.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				_, err := svc.Store(tt.originalURL, uuid.New())
				if err != nil {
					t.Fatalf("Service.Store() error = %v", err)
				}
			}
			err := svc.DeleteByOriginalURL(tt.originalURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.DeleteByOriginalURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_DeleteByShortCode(t *testing.T) {
	svc, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	tests := []struct {
		name        string
		originalURL string
		shortCode   string
		wantErr     bool
	}{
		{"Stored url is deleted", "https://google.com", "", false},
		{"Not stored url", "https://google.com", "notfound", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				urlInserted, err := svc.Store(tt.originalURL, uuid.New())
				if err != nil {
					t.Fatalf("Service.Store() error = %v", err)
				}
				tt.shortCode = urlInserted.ShortCode
			}
			err := svc.DeleteByShortCode(tt.shortCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.DeleteByShortCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_ValidateURL(t *testing.T) {
	svc, err := provideService()
	if err != nil {
		t.Fatalf("provideService() error = %v", err)
	}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https url", "https://google.com", false},
		{"valid http url", "http://google.com", false},
		{"valid url with path", "https://google.com/path", false},
		{"valid url with query", "https://google.com?query=hello", false},
		{"invalid url", "not a url", true},
		{"empty url", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

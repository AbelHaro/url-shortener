package url

import (
	"context"
	"errors"
	"testing"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	idsRangesRepository "github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	counterService "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	idsRangesService "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	"github.com/google/uuid"
)

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
	idsRangesRepository := idsRangesRepository.NewMockRepository()
	idsRangesService := idsRangesService.NewService(idsRangesRepository)
	counterService, err := counterService.NewService(idsRangesService)

	if err != nil {
		return nil, err
	}

	return NewService(repo, counterService), nil
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

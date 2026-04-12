package url

import (
	"testing"

	idsRangesRepository "github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	counterService "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	idsRangesService "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	"github.com/google/uuid"
)

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
			_, err := svc.Store(tt.originalURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
			urlInserted, err := svc.Store(tt.originalURL)
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
			urlInserted, err := svc.Store(tt.originalURL)

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
			urlInserted, err := svc.Store(tt.originalURL)

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

			if urlFound != urlInserted {
				t.Errorf("Service.FindByOriginalURL() = %v, want %v", urlFound, urlInserted)
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
				urlInserted, err := svc.Store(tt.originalURL)
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
				_, err := svc.Store(tt.originalURL)
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
				urlInserted, err := svc.Store(tt.originalURL)
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

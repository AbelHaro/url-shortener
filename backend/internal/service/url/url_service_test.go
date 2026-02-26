package url

import (
	"testing"

	counterRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/counter"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	"github.com/google/uuid"
)

func provideURLService() (*URLService, error) {
	repo := url.NewMockURLRepository()
	counterRepoInstance := counterRepo.NewMockCounterRepository()
	counterSvcInstance, err := counterSvc.NewCounterService(counterRepoInstance)
	if err != nil {
		return nil, err
	}
	return NewURLService(repo, counterSvcInstance), nil
}

func TestURLService_Store(t *testing.T) {
	svc, err := provideURLService()
	if err != nil {
		t.Fatalf("provideURLService() error = %v", err)
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
				t.Errorf("URLService.Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestURLService_DeleteById(t *testing.T) {
	svc, err := provideURLService()
	if err != nil {
		t.Fatalf("provideURLService() error = %v", err)
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
					t.Errorf("URLService.Store() error = %v, wantErr %v", err, tt.wantErr)
				}
				tt.id = urlInserted.ID
			}
			err := svc.DeleteByID(tt.id.String())

			if (err != nil) != tt.wantErr {
				t.Errorf("URLService.DeleteById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

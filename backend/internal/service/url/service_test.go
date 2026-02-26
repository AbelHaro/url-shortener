package url

import (
	"testing"

	counterRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/counter"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	"github.com/google/uuid"
)

func provideService() (*Service, error) {
	repo := url.NewMockRepository()
	counterRepoInstance := counterRepo.NewMockRepository()
	counterSvcInstance, err := counterSvc.NewService(counterRepoInstance)
	if err != nil {
		return nil, err
	}
	return NewService(repo, counterSvcInstance), nil
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

func TestService_DeleteById(t *testing.T) {
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

package idsranges

import (
	"context"
	"fmt"
	"testing"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgresRepository_UsesNumericRangeOrder(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:18",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_PASSWORD": "password",
				"POSTGRES_DB":       "url_shortener_test",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start PostgreSQL container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("terminate PostgreSQL container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("get PostgreSQL host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("get PostgreSQL port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://postgres:password@%s:%s/url_shortener_test?sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL connection: %v", err)
	}
	if err := db.AutoMigrate(&domain.IDsRange{}); err != nil {
		t.Fatalf("migrate IDsRange: %v", err)
	}

	repository := NewPostgresRepository(db)

	t.Run("allocation ignores UUID order", func(t *testing.T) {
		ranges := []domain.IDsRange{
			{
				ID:    uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
				Start: 0,
				Last:  1000,
			},
			{
				ID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Start: 1000,
				Last:  2000,
			},
		}
		if err := db.Create(&ranges).Error; err != nil {
			t.Fatalf("seed ranges: %v", err)
		}

		allocated, err := repository.AllocateNewRange()
		if err != nil {
			t.Fatalf("AllocateNewRange() error = %v", err)
		}
		if allocated.Start != 2000 || allocated.Last != 3000 {
			t.Errorf("allocated range = [%d, %d), want [2000, 3000)", allocated.Start, allocated.Last)
		}
	})

	t.Run("active lookup prefers consumed duplicate", func(t *testing.T) {
		duplicate := domain.IDsRange{
			ID:            uuid.New(),
			Start:         2000,
			Last:          3000,
			CurrentOffset: 500,
		}
		if err := db.Create(&duplicate).Error; err != nil {
			t.Fatalf("seed duplicate range: %v", err)
		}

		active, err := repository.GetActiveRange()
		if err != nil {
			t.Fatalf("GetActiveRange() error = %v", err)
		}
		if active.CurrentOffset != 500 {
			t.Errorf("CurrentOffset = %d, want 500", active.CurrentOffset)
		}
	})
}

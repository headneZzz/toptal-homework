package integration_test

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"log"
	"log/slog"
	"sync"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"toptal/internal/app/config"
	"toptal/internal/app/repository"
	"toptal/internal/pkg/pg"
)

func TestConcurrentPurchases(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal("Failed to start container:", err)
	}
	defer postgresC.Terminate(ctx)

	host, err := postgresC.Host(ctx)
	if err != nil {
		t.Fatal("Failed to get container host:", err)
	}
	mappedPort, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal("Failed to get container port:", err)
	}

	dbConfig := config.DatabaseConfig{
		User:         "test",
		Password:     "test",
		Host:         host,
		Port:         mappedPort.Port(),
		Name:         "testdb",
		SSLMode:      "disable",
		MaxLifetime:  5 * time.Minute,
		MaxIdleConns: 5,
		MaxOpenConns: 10,
	}

	db, err := pg.Connect(dbConfig)
	if err != nil {
		t.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	n := 10
	setupTestData(dbConfig.DSN(), db, n)

	cartCfg := &config.CartConfig{
		CleanupInterval: time.Minute,
		ExpiryTime:      30 * time.Minute,
	}

	var wg sync.WaitGroup
	results := make([]error, n)
	start := time.Now()
	t.Log("Starting concurrent purchases...")

	wg.Add(n)
	for i := 1; i <= n; i++ {
		userID := i
		go func(id int) {
			tmpDb, err := pg.Connect(dbConfig)
			if err != nil {
				t.Errorf("Failed to connect to database: %v", err)
				wg.Done()
				return
			}
			defer tmpDb.Close()
			repo := repository.NewCartRepository(tmpDb, cartCfg)
			t.Logf("Starting purchasing userId: %d", id)
			err = repo.Purchase(context.Background(), userID)
			if err != nil {
				t.Logf("Purchase error userId: %d error: %v", id, err)
			} else {
				t.Logf("Purchase successful userId: %d", id)
			}
			results[id-1] = err
			wg.Done()
		}(userID)
	}
	wg.Wait()

	successCount := 0
	failCount := 0
	for i, err := range results {
		if err == nil {
			successCount++
			t.Logf("User %d: Success", i+1)
		} else {
			failCount++
			t.Logf("User %d: Failed - %v", i+1, err)
		}
	}
	t.Log("Finished concurrent purchases")
	t.Logf("Total time (seconds): %f", time.Since(start).Seconds())
	t.Logf("Successful purchases count: %d", successCount)
	t.Logf("Failed purchases count: %d", failCount)
}

func setupTestData(dsn string, db *pg.DB, n int) {
	slog.Info("Running migrations...")
	m, err := migrate.New("file://../migrations", dsn)
	if err != nil {
		log.Fatalf("failed to init migrations: %s", err.Error())
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("failed to up migrations: %s", err.Error())
		}
	}
	slog.Info("Migrations applied successfully")
	ids := make([]int, n)
	for i := range ids {
		ids[i] = i + 1
	}

	// Insert test categories.
	db.MustExec("INSERT INTO categories (id, name) VALUES (1, 'Test Category')")

	// Insert test books.
	db.MustExec("INSERT INTO books (id, title, year, author, price, stock, category_id) VALUES (1, 'Test Book', 2025, 'Test Author', 100, 5, 1)")

	// Insert test users.
	for _, id := range ids {
		db.MustExec("INSERT INTO users (id, username, password_hash, admin) VALUES ($1, $2, 'hash', false)", id, id)
		db.MustExec("INSERT INTO cart (id, user_id, updated_at) VALUES ($1, $2, now())", id, id)
		db.MustExec("INSERT INTO cart_items (cart_id, book_id, updated_at) VALUES ($1, 1, now())", id)
	}

	log.Printf("Initial book stock %d", 5)
}

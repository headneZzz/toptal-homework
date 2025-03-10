package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"toptal/internal/app/config"
	"toptal/internal/app/repository"
	"toptal/internal/pkg/pg"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	db, err := pg.Connect(cfg.DB)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	cartCfg := &config.CartConfig{
		CleanupInterval: time.Minute,
		ExpiryTime:      30 * time.Minute,
	}

	repo := repository.NewCartRepository(db, cartCfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	n := 10

	setupTestData(db, n)

	var wg sync.WaitGroup
	results := make([]error, n)

	start := time.Now()
	fmt.Println("Starting concurrent purchases...")

	wg.Add(n)
	for i := 1; i <= n; i++ {
		userID := i
		go func(id int) {
			fmt.Printf("Starting purchasing userId: %d \n", id)
			defer wg.Done()
			err := repo.Purchase(context.Background(), userID)
			if err != nil {
				fmt.Printf("Purchase error userId: %d error: %s \n", id, err)
			} else {
				fmt.Printf("Purchase successful userId: %d \n", id)
			}
			results[id-1] = err
		}(userID)
	}

	fmt.Println("Start")
	wg.Wait()

	successCount := 0
	failCount := 0
	for i, err := range results {
		if err == nil {
			successCount++
			fmt.Printf("User %d: Success\n", i+1)
		} else {
			failCount++
			fmt.Printf("User %d: Failed - %v\n", i+1, err)
		}
	}

	fmt.Println("Finished concurrent purchases")

	fmt.Println("Results time seconds", time.Since(start).Seconds())
	fmt.Println("Successful purchases", "count", successCount)
	fmt.Println("Failed purchases", "count", failCount)
}

func setupTestData(db *pg.DB, n int) {
	ids := make([]int, n)
	for i := range ids {
		ids[i] = i + 1
	}
	query := `DELETE FROM cart WHERE user_id IN (?)`
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return
	}
	query = db.Rebind(query)
	db.MustExec(query, args...)

	db.MustExec("UPDATE books SET stock = 5 WHERE id = 1")

	for i := 1; i <= n; i++ {
		db.MustExec("INSERT INTO cart (user_id, book_id, updated_at) VALUES ($1, 1, now())", i)
	}

	var stock int
	db.GetContext(context.Background(), &stock, "SELECT stock FROM books WHERE id = 1")
	fmt.Println("Initial book stock", stock)
}

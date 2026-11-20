package repository

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func TestOrderStatusConcurrentUpdates(t *testing.T) {
	dsn := os.Getenv("BAR108_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set BAR108_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	ctx := context.Background()
	if err := conn.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(conn, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}

	var userID, orderID int32
	err = conn.QueryRowContext(ctx, `INSERT INTO users(name, phone, email, password_hash)
		VALUES ('Integration User', '+79990000001', 'integration-order@bar108.test', 'test') RETURNING id`).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = conn.Exec(`DELETE FROM orders WHERE id = $1`, orderID) })
	t.Cleanup(func() { _, _ = conn.Exec(`DELETE FROM users WHERE id = $1`, userID) })
	err = conn.QueryRowContext(ctx, `INSERT INTO orders(user_id, status, total_price, final_price)
		VALUES ($1, 'pending', 100, 100) RETURNING id`, userID).Scan(&orderID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = conn.Exec(`DELETE FROM order_status_history WHERE order_id = $1`, orderID) })

	repo := NewOrderRepository(conn)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.UpdateOrderStatus(ctx, orderID, "confirmed")
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, apperror.ErrOrderConflict):
			conflicts++
		default:
			t.Fatalf("unexpected update error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("got %d successes and %d conflicts; want one of each", successes, conflicts)
	}
	order, err := db.New(conn).GetOrderByID(ctx, orderID)
	if err != nil || order.Status != "confirmed" {
		t.Fatalf("order status = %q, err = %v", order.Status, err)
	}
}

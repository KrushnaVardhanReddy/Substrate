//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://postgres:postgres@127.0.0.1:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := SetupP17Database(pool); err != nil {
		fmt.Println("Seed failed:", err)
		os.Exit(1)
	}
	fmt.Println("P17 seeded successfully")
}

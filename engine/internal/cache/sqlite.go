package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	GlobalCache *Cache
	once        sync.Once
)

type Cache struct {
	DB *sql.DB
}

type GraphNode struct {
	Repo     string `json:"repo"`
	Type     string `json:"type"`
}

type GraphEdge struct {
	Provider string `json:"provider"`
	Consumer string `json:"consumer"`
	Status   string `json:"status"`
}

func InitCache(dbPath string) (*Cache, error) {
	if GlobalCache != nil {
		return GlobalCache, nil
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache dir: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schemas (
			repo TEXT PRIMARY KEY,
			content TEXT
		);
		CREATE TABLE IF NOT EXISTS dependencies (
			provider TEXT,
			consumer TEXT,
			status TEXT,
			PRIMARY KEY (provider, consumer)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	GlobalCache = &Cache{DB: db}
	return GlobalCache, nil
}

func (c *Cache) SyncFromRemote(ctx context.Context, apiURL string, token string, org string) error {
	if c.DB == nil {
		return fmt.Errorf("cache db not initialized")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/graph/%s", apiURL, org), nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch graph, status: %d", resp.StatusCode)
	}

	var edges []GraphEdge
	if err := json.NewDecoder(resp.Body).Decode(&edges); err != nil {
		return err
	}

	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM dependencies")
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO dependencies (provider, consumer, status) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range edges {
		_, err = stmt.ExecContext(ctx, e.Provider, e.Consumer, e.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

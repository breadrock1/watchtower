package pg

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"watchtower/internal/application/dto"

	_ "github.com/lib/pq"
)

type PgClient struct {
	conn *sqlx.DB
}

func NewPgClient(config *Config) (*PgClient, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DbName, config.SSLMode)

	conn, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to pg: %w", err)
	}

	err = conn.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping pg: %w", err)
	}

	conn.SetMaxOpenConns(3)                  // Maximum number of open connections
	conn.SetMaxIdleConns(3)                  // Maximum number of idle connections
	conn.SetConnMaxLifetime(5 * time.Minute) // Maximum amount of time a connection may be reused

	return &PgClient{conn: conn}, nil
}

func (pc *PgClient) LoadAllWatcherDirs(ctx context.Context) ([]dto.Directory, error) {
	query := `SELECT bucket, path, created_at FROM watched_directories`

	rows, err := pc.conn.QueryxContext(ctx, query)
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to select watched dirs: %w", err)
	}

	var allDirs []dto.Directory
	for rows.Next() {
		var dir dto.Directory
		err = rows.StructScan(&dir)
		if err != nil {
			log.Printf("failed to scan directory rows: %v", err)
			continue
		}

		allDirs = append(allDirs, dir)
	}

	return allDirs, nil
}

func (pc *PgClient) StoreWatcherDir(ctx context.Context, dir dto.Directory) error {
	query := `
		INSERT INTO watched_directories (bucket, path, created_at) 
		VALUES (:first_name, :last_name, :email)
	`

	_, err := pc.conn.NamedExecContext(ctx, query, dir)
	if err != nil {
		return fmt.Errorf("failed to store watcher dir: %w", err)
	}

	return nil
}

func (pc *PgClient) DeleteWatcherDir(ctx context.Context, bucket, filePath string) error {
	query := `DELETE FROM watched_directories WHERE bucket=$1 AND path=$2`
	_, err := pc.conn.ExecContext(ctx, query, bucket, filePath)
	if err != nil {
		return fmt.Errorf("failed to store watcher dir: %w", err)
	}

	return nil
}

func (pc *PgClient) DeleteAllWatcherDirs(ctx context.Context) error {
	query := `DELETE FROM watched_directories`
	_, err := pc.conn.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to store watcher dir: %w", err)
	}

	return nil
}

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

	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to pg: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping pg: %w", err)
	}

	db.SetMaxOpenConns(3)                  // Maximum number of open connections
	db.SetMaxIdleConns(3)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum amount of time a connection may be reused

	return &PgClient{conn: db}, nil
}

func (pc *PgClient) LoadAllWatcherDirs(ctx context.Context) ([]dto.Directory, error) {
	query := `SELECT bucket, path, created_at FROM watched_directories`

	rows, err := pc.conn.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to select watched dirs: %w", err)
	}

	var allDirs []dto.Directory
	for rows.Next() {
		var p dto.Directory
		err = rows.StructScan(&p)
		if err != nil {
			log.Printf("failed to scan directory rows: %v", err)
			continue
		}

		allDirs = append(allDirs, p)
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

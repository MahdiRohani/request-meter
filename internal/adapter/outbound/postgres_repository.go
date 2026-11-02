package outbound

import (
	"context"
	"database/sql"
	"log"

	"github.com/MahdiRohani/request-meter/internal/domain"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbURL string) *PostgresRepository {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is not reachable: %v", err)
	}

	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) Close(ctx context.Context) error {
	log.Println("Closing postgres connection...")
	return p.db.Close()
}

func (p *PostgresRepository) SaveBatch(ctx context.Context, counters []*domain.Counter) error {
	if len(counters) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	defer tx.Rollback()

	query := `
		INSERT INTO counters (name, count)
		VALUES ($1, $2)
		ON CONFLICT (name) 
		DO UPDATE SET count = EXCLUDED.count; 
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, c := range counters {
		_, err := stmt.ExecContext(ctx, c.Name, c.Count)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (p *PostgresRepository) GetAll(ctx context.Context) []*domain.Counter {
	rows, err := p.db.QueryContext(ctx, "SELECT name, count FROM counters")
	if err != nil {
		log.Printf("Failed to fetch counters: %v", err)
		return nil
	}
	defer rows.Close()

	var snapshot []*domain.Counter
	for rows.Next() {
		var c domain.Counter
		if err := rows.Scan(&c.Name, &c.Count); err == nil {
			snapshot = append(snapshot, &c)
		} else {
			log.Printf("Error scanning row: %v", err)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		return nil
	}
	return snapshot
}

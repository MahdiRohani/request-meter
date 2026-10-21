package outbound

import (
	"database/sql"
	"log"

	"github.com/MahdiRohani/request-meter/internal/core/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbURL string) *PostgresRepository {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS counters (
			name VARCHAR(255) PRIMARY KEY,
			count BIGINT NOT NULL
		);
	`)

	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) SaveBatch(counters []*domain.Counter) error {
	if len(counters) == 0 {
		return nil
	}

	tx, err := p.db.Begin()
	defer tx.Rollback()

	query := `
		INSERT INTO counters (name, count)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE 
		SET count = counters.count + EXCLUDED.count;
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, c := range counters {
		_, err := tx.Exec(c.Name, c.Count)
		if err != nil {
			return nil
		}
	}
	return tx.Commit()
}

func (p *PostgresRepository) GetAll() []*domain.Counter {
	rows, err := p.db.Query("SELECT name, count FROM counters")
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
		}
	}
	return snapshot
}

package shorty

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"contrib.go.opencensus.io/integrations/ocsql"
	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

// Persistence represents the database backend.
type Persistence struct {
	config *Config
	mu     *sync.Mutex
	db     *sql.DB
}

// Write creates a new entry in the database.
func (p *Persistence) Write(ctx context.Context, s *Shortened) error {
	var tx *sql.Tx
	var stmt *sql.Stmt
	var err error

	p.mu.Lock()
	defer p.mu.Unlock()

	query := `
INSERT INTO shorty(short, url, created_at)
VALUES (?, ?, ?)`

	if tx, err = p.db.Begin(); err != nil {
		return err
	}
	if stmt, err = tx.PrepareContext(ctx, query); err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.ExecContext(ctx, s.Short, s.URL, s.CreatedAt); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Read database entry based on its short string, unique in the database.
func (p *Persistence) Read(ctx context.Context, short string) (*Shortened, error) {
	var rows *sql.Rows
	var err error

	query := `
SELECT short, url, created_at
FROM shorty
WHERE short = ?`

	if rows, err = p.db.QueryContext(ctx, query, short); err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	s := &Shortened{}
	if err = rows.Scan(&s.Short, &s.URL, &s.CreatedAt); err != nil {
		return nil, err
	}

	return s, nil
}

// addSchema create shorty table, if not present yet.
func (p *Persistence) addSchema() error {
	log.Printf("Creating 'shorty' table, if not present.")
	createTable := `
CREATE TABLE IF NOT EXISTS shorty (
	short  		TEXT NOT NULL,
	url 	    TEXT NOT NULL,
	created_at 	INTEGER NOT NULL,
	PRIMARY KEY (short)
)`
	if _, err := p.db.Exec(createTable); err != nil {
		return err
	}
	return nil
}

// IsErrNoRows assert if error is about no rows found.
func (p *Persistence) IsErrNoRows(err error) bool {
	return sql.ErrNoRows == err
}

// IsErrUniqueConstraint check if informed error is about violation of unique constraint.
func (p *Persistence) IsErrUniqueConstraint(err error) bool {
	return strings.HasPrefix(err.Error(), "UNIQUE constraint failed")
}

// Close terminate the connection with database.
func (p *Persistence) Close() {
	if err := p.db.Close(); err != nil {
		log.Printf("Error on closing database connection: '%s'", err)
	}
}

// NewPersistence creates a new persistence instance, opens database connection and add schema.
func NewPersistence(config *Config) (*Persistence, error) {
	var driverName string
	var err error

	if driverName, err = ocsql.Register("sqlite3", ocsql.WithAllTraceOptions()); err != nil {
		log.Fatalf("failed to register ocsql driver: %v\n", err)
		return nil, err
	}
	ocsql.RegisterAllViews()

	var connStr string
	if config.DatabaseFile == "" {
		log.Printf("Starting a in-memory database...")
		connStr = fmt.Sprintf("file::memory:?cache=shared&%s", config.SQLiteFlags)
	} else {
		connStr = fmt.Sprintf("%s?%s", config.DatabaseFile, config.SQLiteFlags)
	}
	log.Printf("New database connection '%s'", connStr)

	p := &Persistence{config: config, mu: &sync.Mutex{}}

	if p.db, err = sql.Open(driverName, connStr); err != nil {
		return nil, err
	}
	if err = p.addSchema(); err != nil {
		return nil, err
	}

	return p, nil
}

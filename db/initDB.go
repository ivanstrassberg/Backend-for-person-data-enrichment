package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/source/file"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewConnString() string {

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	host := os.Getenv("DB_HOST")
	sslmode := os.Getenv("DB_SSLMODE")
	return fmt.Sprintf("host=%s user=%s port=%s dbname=%s password=%s sslmode=%s", host, user, port, dbname, password, sslmode)
}

func RunMigrations(db *sql.DB) error {
	_, callerPath, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filepath.Dir(callerPath)), "db", "migrations")

	goose.SetBaseFS(os.DirFS(migrationsPath))

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose up failed: %w", err)
	}

	log.Println("migration success")
	return nil
}
func NewPostgresStorage() (*PostgresStorage, error) {
	connStr := NewConnString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

// ostgres=# create user em_user with PASSWORD 'effectivemobile';
// CREATE ROLE
// postgres=# GRANT ALL PRIVILEGES ON DATABASE effective_mobile TO em_user;
// GRANTostgres=# create user em_user with PASSWORD 'effectivemobile';
// CREATE ROLE
// postgres=# GRANT ALL PRIVILEGES ON DATABASE effective_mobile TO em_user;
// GRANT

// ALTER DEFAULT PRIVILEGES IN SCHEMA public
// GRANT ALL PRIVILEGES ON TABLES TO em_user;
// GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO em_user;
// GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO em_user;
// GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO em_user;
// -- GRANT ALL PRIVILEGES ON SCHEMA public TO em_user;

// -- ALTER DEFAULT PRIVILEGES IN SCHEMA public
// -- GRANT ALL PRIVILEGES ON TABLES TO em_user;

// -- ALTER DEFAULT PRIVILEGES IN SCHEMA public
// -- GRANT ALL PRIVILEGES ON SEQUENCES TO em_user;

// -- GRANT CONNECT ON DATABASE effective_mobile to em_user;
// -- GRANT CREATE ON DATABASE effective_mobile to em_user;

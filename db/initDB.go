package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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

	goose.SetBaseFS(os.DirFS("./db/migrations"))

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
	// Init(db, dbname, user, password)
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func Init(db *sql.DB, dbName, userName, userPassword string) error {
	statements := []string{
		fmt.Sprintf(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s') THEN
					CREATE ROLE %s WITH LOGIN PASSWORD '%s';
				END IF;
			END$$;`, userName, userName, userPassword),

		fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s;`, dbName, userName),
		fmt.Sprintf(`GRANT CONNECT ON DATABASE %s TO %s;`, dbName, userName),
		fmt.Sprintf(`GRANT CREATE ON DATABASE %s TO %s;`, dbName, userName),

		fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %s;`, userName),
		fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO %s;`, userName),
		fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO %s;`, userName),

		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO %s;`, userName),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO %s;`, userName),
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("error executing statement: %v\nSQL: %s", err, stmt)
		}
	}
	return nil

}

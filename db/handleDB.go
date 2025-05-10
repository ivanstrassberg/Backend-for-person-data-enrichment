package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Storage interface {
	CreatePerson(string, string, string, int, string, string) error
	DeletePerson(int) error
	UpdatePersonEnrich(int, string, string, string, int, string, string) (string, error)
	UpdatePerson(int, string, string, string, int, string, string) (string, error)
	CheckName(int) (string, error)
}

func (s *PostgresStorage) CreatePerson(name, surname, patronymic string, age int, gender, nationality string) error {

	query := `
		insert into em_people1 
		(fname, surname, patronymic, age, nationality, gender) 
		values ($1, $2, $3, $4, $5, $6)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query,
		name,
		surname,
		patronymic,
		age,
		nationality,
		gender,
	)

	if err != nil {
		return fmt.Errorf("failed to create person: %w", err)
	}

	return nil
}

func (s *PostgresStorage) DeletePerson(id int) error {
	query := `DELETE FROM em_people1 WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete person: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("person with id %d does not exist", id)
	}

	return nil
}

func (s *PostgresStorage) UpdatePersonEnrich(id int, name, surname, patronymic string, age int, gender, nationality string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `
        update em_people1 set
            fname = $1,
            surname = $2,
            patronymic = $3,
            age = $4,
            gender = $5,
            nationality = $6
        where id = $7
    `, name, surname, patronymic, age, gender, nationality, id)

	if err != nil {
		return fmt.Errorf("failed to update person: %w", err)
	}

	return nil
}

func (s *PostgresStorage) CheckName(id int) (string, error) {
	query := `select fname from em_people1 where id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var currentName string
	err := s.db.QueryRowContext(ctx, query, id).Scan(&currentName)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("person with id %d not found", id)
		}
		return "", fmt.Errorf("failed to get person: %w", err)
	}
	return currentName, nil
}

func (s *PostgresStorage) UpdatePerson(id int, name, surname, patronymic string, age int, gender, nationality string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `
        update em_people1 set
            fname = $1,
            surname = $2,
            patronymic = $3,
            age = $4,
            gender = $5,
            nationality = $6
        where id = $7
    `, name, surname, patronymic, age, gender, nationality, id)

	if err != nil {
		return fmt.Errorf("failed to update person: %w", err)
	}

	return nil
}

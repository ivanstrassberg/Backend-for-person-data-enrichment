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
	UpdatePersonPatch(int, string, string, string, int, string, string) (string, error)
	CheckName(int) (string, error)
}

type Person struct {
	ID          int
	Name        string
	Surname     string
	Patronymic  string
	Age         int
	Gender      string
	Nationality string
}

func (s *PostgresStorage) GetPeopleWithPagination(fname, surname, patronymic string, age int, nationality, gender string, limit, offset int) ([]Person, int, error) {
	query := `SELECT id, fname, surname, patronymic, age, nationality, gender FROM em_people1 WHERE 1=1`
	countQuery := `SELECT count(*) FROM em_people1 WHERE 1=1`
	var args []interface{}
	var countArgs []interface{}
	argCount := 1

	addFilter := func(field, value string) {
		query += fmt.Sprintf(" AND %s ILIKE '%%' || $%d || '%%'", field, argCount)
		countQuery += fmt.Sprintf(" AND %s ILIKE '%%' || $%d || '%%'", field, argCount)
		args = append(args, value)
		countArgs = append(countArgs, value)
		argCount++
	}

	if fname != "" {
		addFilter("fname", fname)
	}
	if surname != "" {
		addFilter("surname", surname)
	}
	if patronymic != "" {
		addFilter("patronymic", patronymic)
	}
	if age > 0 {
		query += fmt.Sprintf(" AND age = $%d", argCount)
		countQuery += fmt.Sprintf(" AND age = $%d", argCount)
		args = append(args, age)
		countArgs = append(countArgs, age)
		argCount++
	}
	if nationality != "" {
		addFilter("nationality", nationality)
	}
	if gender != "" {
		addFilter("gender", gender)
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var people []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Surname, &p.Patronymic, &p.Age, &p.Nationality, &p.Gender); err != nil {
			return nil, 0, err
		}
		people = append(people, p)
	}

	var total int
	if err := s.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return people, total, nil
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

func (s *PostgresStorage) UpdatePersonPatch(id int, name, surname, patronymic string, age int, gender, nationality string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := "UPDATE em_people1 SET"
	args := []interface{}{}
	argPos := 1

	addField := func(field string, value interface{}, omitEmpty bool) {
		if !omitEmpty || !isZero(value) {
			if len(args) > 0 {
				query += ","
			}
			query += fmt.Sprintf(" %s = $%d", field, argPos)
			args = append(args, value)
			argPos++
		}
	}

	addField("fname", name, true)
	addField("surname", surname, true)
	addField("patronymic", patronymic, true)
	addField("age", age, true)
	addField("gender", gender, true)
	addField("nationality", nationality, true)

	if len(args) == 0 {
		return nil
	}

	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, id)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update person: %w", err)
	}

	return nil
}

func isZero(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	case int:
		return v == 0
	default:
		return value == nil
	}
}

package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" //GOLINT says I have to explain why I have this here... It's neater here.
)

// DBSession ...
type DBSession struct {
	db *sql.DB
}

// New session
func New(user, pass, addr, name string) (*DBSession, error) {

	address := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", user, pass, addr, name)
	//address := fmt.Sprintf("postgres://%s/%s?sslmode=disable", addr, name)

	session, err := sql.Open("postgres", address)
	if err != nil {
		return nil, err
	}

	if err = session.Ping(); err != nil {
		return nil, err
	}

	return &DBSession{db: session}, nil
}

// Execute records
func (s *DBSession) Execute(statement string, a ...interface{}) error {

	_, err := s.db.Exec(statement, a...)
	if err != nil {
		return err
	}
	return nil
}

// Query records
func (s *DBSession) Query(statement string, a ...interface{}) (*sql.Rows, error) {
	rows, err := s.db.Query(statement, a...)
	if err != nil {
		return nil, err
	}
	return rows, nil

}

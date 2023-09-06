package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
	"time"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123456"
	dbname   = "postgres"
)

type PostgresDB struct {
	Dbms
	db *sql.DB
}

func InitPostgresql() (*PostgresDB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return &PostgresDB{}, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			subject TEXT,
			body TEXT,
-- 			publishedAt TIMESTAMP,
			expiration TIMESTAMP
		);
	`)

	if err != nil {
		err = db.Close()
		return &PostgresDB{}, err
	}

	return &PostgresDB{
		db: db,
	}, err
}

func (ps *PostgresDB) SendMessage(message model.Message, subject string) (int, error) {
	row := ps.db.QueryRow(`
		INSERT INTO messages (subject, body, expiration)
		VALUES ($1, $2, $3)
		RETURNING id
		`, subject, message.Body, time.Now().Add(message.Expiration))

	var id int64
	err := row.Scan(&id)
	if err != nil {
		return -1, err
	}

	return int(id), nil
}

func (ps *PostgresDB) FetchMessage(messageId int, subject string) (model.Message, error) {
	var message model.Message
	var expiration time.Time

	err := ps.db.QueryRow(`
		SELECT id, body, expiration
		FROM messages
		WHERE subject = $1 AND id = $2
	`, subject, messageId).Scan(&message.Id, &message.Body, &expiration)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Message{}, broker.ErrInvalidID
		}
		return model.Message{}, err
	}

	if time.Since(expiration) > 0 {
		return model.Message{}, broker.ErrExpiredID
	}

	return message, nil
}

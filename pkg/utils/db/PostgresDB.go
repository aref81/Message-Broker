package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"log"
	"sync"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
	"time"
)

const (
	host          = "db"
	port          = 5432
	user          = "postgres"
	password      = "123456"
	dbname        = "postgres"
	maxConnection = 100

	maxBatchSize = 30000
)

type PostgresDB struct {
	Dbms
	db   *sql.DB
	pool *pgxpool.Pool

	mu               sync.Mutex
	maxBatchSize     int
	accumulatedCount int
	messages         []model.Message
	subjects         []string
}

//func InitPostgresql() (*PostgresDB, error) {
//	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable pool_max_conns=%d",
//		host, port, user, password, dbname, maxConnection)
//
//	db, err := sql.Open("postgres", psqlInfo)
//	if err != nil {
//		return &PostgresDB{}, err
//	}
//
//	_, err = db.Exec(`
//		CREATE TABLE IF NOT EXISTS messages (
//			id SERIAL PRIMARY KEY,
//			subject TEXT,
//			body TEXT,
//-- 			publishedAt TIMESTAMP,
//			expiration TIMESTAMP
//		);
//	`)
//
//	if err != nil {
//		err = db.Close()
//		return &PostgresDB{}, err
//	}
//
//	return &PostgresDB{
//		db: db,
//	}, err
//}

func InitPostgresql() (*PostgresDB, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable pool_max_conns=%d",
		host, port, user, password, dbname, maxConnection))
	if err != nil {
		return &PostgresDB{}, err
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return &PostgresDB{}, err
	}

	_, err = pool.Exec(context.Background(), `
	   CREATE TABLE IF NOT EXISTS messages (
	       id SERIAL PRIMARY KEY,
	       subject TEXT,
	       body TEXT,
	       expiration TIMESTAMP
	   );
	`)

	if err != nil {
		pool.Close()
		return &PostgresDB{}, err
	}

	//	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	//		host, port, user, password, dbname)
	//
	//	db, err := sql.Open("postgres", psqlInfo)
	//	if err != nil {
	//		return &PostgresDB{}, err
	//	}
	//
	//	_, err = db.Exec(`
	//		CREATE TABLE IF NOT EXISTS messages (
	//			id SERIAL PRIMARY KEY,
	//			subject TEXT,
	//			body TEXT,
	//-- 			publishedAt TIMESTAMP,
	//			expiration TIMESTAMP
	//		);
	//	`)
	//
	//	if err != nil {
	//		err = db.Close()
	//		return &PostgresDB{}, err
	//	}

	return &PostgresDB{
		pool:         pool,
		mu:           sync.Mutex{},
		maxBatchSize: maxBatchSize,
		messages:     make([]model.Message, 0),
		subjects:     make([]string, 0),
	}, err
}

//func (ps *PostgresDB) SendMessage(message model.Message, subject string) (int, error) {
//	_, _ = ps.db.Exec(`
//		INSERT INTO messages (subject, body, expiration)
//		VALUES ($1, $2, $3)
//		`, subject, message.Body, time.Now())
//
//	//var id int64
//	//err := row.Scan(&id)
//	//if err != nil {
//	//	return -1, err
//	//}
//
//	return 0, nil
//}

//func (ps *PostgresDB) SendMessage(message model.Message, subject string) (int, error) {
//	_, err := ps.pool.Exec(context.Background(), `
//        INSERT INTO messages (subject, body, expiration)
//        VALUES ($1, $2, $3)
//    `, subject, message.Body, time.Now())
//
//	if err != nil {
//		return -1, err
//	}
//
//	return 0, nil
//}

func (ps *PostgresDB) SendMessage(message model.Message, subject string) (int, error) {
	ps.messages = append(ps.messages, message)
	ps.subjects = append(ps.subjects, subject)
	ps.accumulatedCount++

	if ps.accumulatedCount >= maxBatchSize {
		messagesCp := ps.messages
		subjectsCP := ps.subjects

		count := ps.accumulatedCount

		ps.messages = make([]model.Message, 0)
		ps.subjects = make([]string, 0)
		ps.accumulatedCount = 0

		go func() {
			err := ps.executeBatchQuery(messagesCp, subjectsCP, count)
			if err != nil {
				logrus.Fatalf("Postgres Error : %s", err.Error())
			}
		}()
	}

	return 0, nil
}

func (ps *PostgresDB) executeBatchQuery(messages []model.Message, subjects []string, count int) error {

	tx, err := ps.pool.Begin(context.Background())
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	stmt := "INSERT INTO messages (subject, body, expiration) VALUES ($1, $2, $3)"

	for i, msg := range messages {
		_, err := tx.Exec(context.Background(), stmt, subjects[i], msg.Body, time.Now())
		if err != nil {
			log.Fatalf("Error inserting row %d: %v", i, err)
		}
	}
	err = tx.Commit(context.Background())
	if err != nil {
		log.Fatalf("Error committing transaction: %v", err)
	}

	return nil
}

//func (ps *PostgresDB) executeBatchQuery(messages []model.Message, subjects []string, count int) error {
//	tx, err := ps.db.Begin()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer tx.Rollback()
//
//	stmt, err := tx.Prepare("INSERT INTO messages (subject, body, expiration) VALUES ($1, $2, $3)")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stmt.Close()
//
//	// Loop through the data and perform batch inserts
//	for i, msg := range messages {
//		_, err := stmt.Exec(subjects[i], msg.Body, time.Now())
//		if err != nil {
//			log.Fatalf("Error inserting row %d: %v", i, err)
//		}
//
//		// Commit the transaction if we've reached the batch size or processed all data
//		//if (i+1)%maxBatchSize == 0 || (i+1) == len(data) {
//		//	err = tx.Commit()
//		//	if err != nil {
//		//		log.Fatal(err)
//		//	}
//		//
//		//	// Start a new transaction
//		//	tx, err = db.Begin()
//		//	if err != nil {
//		//		log.Fatal(err)
//		//	}
//		//	defer tx.Rollback() // Rollback if there's an error, otherwise, we'll commit at the end
//		//}
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		log.Fatal(err)
//		return err
//	}
//
//	return nil
//}

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

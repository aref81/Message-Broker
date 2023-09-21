package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	_ "strings"
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

	syncInterval = 10 * time.Millisecond
	batchSize    = 200000
)

type PostgresDB struct {
	Dbms
	pool *pgxpool.Pool
	mu   sync.Mutex

	messages [][]interface{}
	ctx      context.Context
}

func InitPostgresql() (*PostgresDB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	poolConfig, err := pgxpool.ParseConfig(psqlInfo)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = maxConnection

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS messages (
			id BIGINT PRIMARY KEY,
			subject TEXT,
			body TEXT,
			expiration TIMESTAMP
		);
	`)

	if err != nil {
		return nil, err
	}

	ps := &PostgresDB{
		pool:     pool,
		mu:       sync.Mutex{},
		messages: make([][]interface{}, 0),
		ctx:      context.Background(),
	}

	//go databaseSyncer(ps)
	return ps, err
}

func (ps *PostgresDB) SendMessage(message model.Message, subject string) (int, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	//postID := int(time.Now().UnixNano())
	postID := GenerateUniqueID()
	message.Id = postID

	ps.messages = append(ps.messages,
		[]interface{}{
			postID,
			subject,
			message.Body,
			time.Now().Add(message.Expiration),
		})

	// batching based on maxSize
	if len(ps.messages) >= batchSize {
		go ps.sync()
	}

	return message.Id, nil
}

func databaseSyncer(ps *PostgresDB) {
	ticker := time.NewTicker(syncInterval)

	for {
		select {
		case <-ps.ctx.Done():
			return
		case <-ticker.C:
			ps.sync()
		}
	}
}

func (ps *PostgresDB) sync() {
	ps.mu.Lock()

	messages := ps.messages
	ps.messages = make([][]interface{}, 0)

	ps.mu.Unlock()

	if len(messages) != 0 {
		//chunks := chunkSlice(messages, batchSize)
		//err := ps.bulkInsert(chunks)
		err := ps.bulkInsert(messages)
		if err != nil {
			logrus.Fatalf("ERROR: cannot insert to database: %s", err)
		}
	}
}

//func (ps *PostgresDB) bulkInsert(messageChunks [][][]interface{}) error {
//	for _, chunk := range messageChunks {
//		valueStrings := make([]string, 0, len(chunk))
//		valueArgs := make([]interface{}, 0, len(chunk)*4)
//		for i, msg := range chunk {
//			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
//			valueArgs = append(valueArgs, msg[0])
//			valueArgs = append(valueArgs, msg[1])
//			valueArgs = append(valueArgs, msg[2])
//			valueArgs = append(valueArgs, msg[3])
//		}
//		stmt := fmt.Sprintf("INSERT INTO messages (id, subject, body, expiration) VALUES %s",
//			strings.Join(valueStrings, ","))
//		_, err := ps.pool.Exec(context.Background(), stmt, valueArgs...)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}

func (ps *PostgresDB) bulkInsert(messages [][]interface{}) error {
	_, err := ps.pool.CopyFrom(ps.ctx,
		pgx.Identifier{"messages"},
		[]string{"id", "subject", "body", "expiration"},
		pgx.CopyFromRows(messages),
	)
	if err != nil {
		logrus.Fatalf("ERROR: cannot insert to postgresql: %s", err.Error())
	}

	return nil
}

func chunkSlice(slice [][]interface{}, chunkSize int) [][][]interface{} {
	var result [][][]interface{}

	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}

	return result
}

func (ps *PostgresDB) FetchMessage(messageId int, subject string) (model.Message, error) {
	var message model.Message
	var expiration time.Time

	err := ps.pool.QueryRow(ps.ctx, `
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

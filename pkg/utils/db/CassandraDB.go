package db

import (
	_ "fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
)

const (
	cassandraHosts    = "localhost"
	cassandraKeyspace = "yes"
)

type CassandraDB struct {
	Dbms
	session *gocql.Session
}

func InitCassandra() (*CassandraDB, error) {
	cluster := gocql.NewCluster(cassandraHosts)
	cluster.Keyspace = cassandraKeyspace
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS messages (
			id BIGINT PRIMARY KEY,
			subject TEXT,
			body TEXT,
			expiration TIMESTAMP
		);
	`

	err = session.Query(createTableQuery).Exec()
	if err != nil {
		session.Close()
		log.Fatal(err)
		return nil, err
	}

	return &CassandraDB{
		session: session,
	}, nil
}

func (c *CassandraDB) SendMessage(message model.Message, subject string) (int, error) {
	id := int(time.Now().UnixNano())

	if err := c.session.Query(`
		INSERT INTO yes.messages (id, subject, body, expiration)
		VALUES (?, ?, ?, ?)
	`, id, subject, message.Body, time.Now().Add(message.Expiration)).Exec(); err != nil {
		return -1, err
	}

	return id, nil
}

func (c *CassandraDB) FetchMessage(messageId int, subject string) (model.Message, error) {
	var message model.Message
	var expiration time.Time

	query := c.session.Query(`
		SELECT id, body, expiration
		FROM yes.messages
		WHERE subject = ? AND id = ?
		ALLOW FILTERING
`, subject, messageId)

	err := query.Scan(&message.Id, &message.Body, &expiration)
	if err != nil {
		if err == gocql.ErrNotFound {
			return model.Message{}, broker.ErrInvalidID
		}
		return model.Message{}, err
	}

	if time.Now().After(expiration) {
		return model.Message{}, broker.ErrExpiredID
	}

	return message, nil
}

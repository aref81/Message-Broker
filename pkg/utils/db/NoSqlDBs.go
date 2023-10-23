package db

import (
	"fmt"
	_ "fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
)

const (
	cassandraHosts    = "db"
	cassandraKeyspace = "yes"

	scyllaHosts    = "db"
	scyllaKeyspace = "yes"
)

type NoSqlDB struct {
	Dbms
	session *gocql.Session
}

func InitCassandra() (*NoSqlDB, error) {
	cluster := gocql.NewCluster(cassandraHosts)

	err := createKeySpace(cluster, cassandraKeyspace)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

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

	return &NoSqlDB{
		session: session,
	}, nil
}

func InitScylla() (*NoSqlDB, error) {
	cluster := gocql.NewCluster(scyllaHosts)

	err := createKeySpace(cluster, scyllaKeyspace)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	cluster.Keyspace = scyllaKeyspace
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

	return &NoSqlDB{
		session: session,
	}, nil
}

func createKeySpace(cluster *gocql.ClusterConfig, keysSpace string) error {
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	replicationStrategy := fmt.Sprintf("{'class':'SimpleStrategy', 'replication_factor':%d}", 1)
	createKeyspaceQuery := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH replication = %s", keysSpace, replicationStrategy)
	if err := session.Query(createKeyspaceQuery).Exec(); err != nil {
		log.Fatal(err)
	}
	session.Close()
	return err
}

func (ns *NoSqlDB) SendMessage(message model.Message, subject string) (int, error) {
	id := int(time.Now().UnixNano())

	if err := ns.session.Query(`
		INSERT INTO yes.messages (id, subject, body, expiration)
		VALUES (?, ?, ?, ?)
	`, id, subject, message.Body, time.Now().Add(message.Expiration)).Exec(); err != nil {
		return -1, err
	}

	return id, nil
}

func (ns *NoSqlDB) FetchMessage(messageId int, subject string) (model.Message, error) {
	var message model.Message
	var expiration time.Time

	query := ns.session.Query(`
		SELECT id, body, expiration
		FROM yes.messages
		WHERE subject = ? AND id = ?
		ALLOW FILTERING
`, subject, messageId) // TODO : Don't Allow Filtering

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

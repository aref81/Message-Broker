package db

import "therealbroker/internal/broker/model"

const (
	INMEM     = 0
	CASSANDRA = 1
	POSTGRES  = 2
	SCYLLA    = 3
)

type Dbms interface {
	Close() error

	//AddSubject(subject model.Subject) (int, error)
	//DeleteSubject(subjectID int) error

	SendMessage(message model.Message, subject string) (int, error)
	FetchMessage(messageId int, subject string) (model.Message, error)
	//DeleteMessage(messageId int) (model.Message, error)
}

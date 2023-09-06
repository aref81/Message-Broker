package db

import "therealbroker/internal/broker/model"

type Dbms interface {
	Close() error

	//AddSubject(subject model.Subject) (int, error)
	//DeleteSubject(subjectID int) error

	SendMessage(message model.Message, subject string) (int, error)
	FetchMessage(messageId int, subject string) (model.Message, error)
	//DeleteMessage(messageId int) (model.Message, error)
}

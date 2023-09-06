package model

import "time"

type Message struct {
	// This parameter is optional. If it's not provided,
	// the Message can't be accessible through Fetch()
	// id is unique per every subject
	Id int // CONSIDER
	// Body of the message
	Body string
	// The time that message can be accessible through Fetch()
	// with the proper Message id
	// 0 when there is no need to keep message ( fire & forget mode )
	Expiration time.Duration
}

//type Pair struct {
//	Message Message
//	Sent    time.Time
//}

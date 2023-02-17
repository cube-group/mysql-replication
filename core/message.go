package core

import "github.com/withlin/canal-go/protocol/entry"

type ReplicationMessage struct {
	LogfileName   string
	LogfileOffset int64
	SchemaName    string
	TableName     string
	EventType     entry.EventType
	Body          map[string]interface{}
	Columns       []string // the column about updated
}

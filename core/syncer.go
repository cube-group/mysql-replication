// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/withlin/canal-go/client"
	pbe "github.com/withlin/canal-go/protocol/entry"
)

type ReplicationDMLHandler func(msg ReplicationMessage)

type ReplicationSyncer struct {
	option     CanalOption
	dmlHandler ReplicationDMLHandler
}

func NewReplicationSyncer(canalOption CanalOption, dmlHandler ReplicationDMLHandler) *ReplicationSyncer {
	if canalOption.SoTimOut == 0 {
		canalOption.SoTimOut = 60000 // default 1min
	}
	if canalOption.IdleTimeOut == 0 {
		canalOption.IdleTimeOut = 60 * 60 * 1000 // default 1hour
	}
	if canalOption.BatchSize == 0 {
		canalOption.BatchSize = 100 // default batchSize
	}
	i := new(ReplicationSyncer)
	i.option = canalOption
	i.dmlHandler = dmlHandler
	return i
}

func (t *ReplicationSyncer) Start() error {
	// 192.168.199.17 替换成你的canal server的地址
	// example 替换成-e canal.destinations=example 你自己定义的名字
	connector := client.NewSimpleCanalConnector(
		t.option.Host,
		t.option.Port,
		t.option.Username,
		t.option.Password,
		t.option.Destination,
		t.option.SoTimOut,
		t.option.IdleTimeOut,
	)
	err := connector.Connect()
	if err != nil {
		return err
	}
	defer connector.DisConnection()

	// https://github.com/alibaba/canal/wiki/AdminGuide
	//mysql 数据解析关注的表，Perl正则表达式.
	//
	//多个正则之间以逗号(,)分隔，转义符需要双斜杠(\\)
	//
	//常见例子：
	//
	//  1.  所有表：.*   or  .*\\..*
	//	2.  canal schema下所有表： canal\\..*
	//	3.  canal下的以canal打头的表：canal\\.canal.*
	//	4.  canal schema下的一张表：canal\\.test1
	//  5.  多个规则组合使用：canal\\..*,mysql.test1,mysql.test2 (逗号分隔)

	err = connector.Subscribe(t.option.Filter)
	if err != nil {
		return err
	}

	for {
		message, err := connector.Get(t.option.BatchSize, nil, nil)
		if err != nil {
			return err
		}
		batchId := message.Id
		if batchId == -1 || len(message.Entries) <= 0 {
			time.Sleep(300 * time.Millisecond)
			continue
		}
		t.printEntry(message.Entries)
	}
}

func (t *ReplicationSyncer) printEntry(entrys []pbe.Entry) {
	for _, entry := range entrys {
		if entry.GetEntryType() == pbe.EntryType_TRANSACTIONBEGIN || entry.GetEntryType() == pbe.EntryType_TRANSACTIONEND {
			continue
		}
		rowChange := new(pbe.RowChange)

		err := proto.Unmarshal(entry.GetStoreValue(), rowChange)
		if err != nil {
			continue
		}
		if rowChange != nil {
			header := entry.GetHeader()
			for _, rowData := range rowChange.GetRowDatas() {
				t.send(header, rowData)
			}
		}
	}
}

func (t *ReplicationSyncer) send(header *pbe.Header, rowData *pbe.RowData) {
	msg := ReplicationMessage{
		LogfileName:   header.GetLogfileName(),
		LogfileOffset: header.GetLogfileOffset(),
		SchemaName:    header.GetSchemaName(),
		TableName:     header.GetTableName(),
		EventType:     header.GetEventType(),
	}
	switch msg.EventType {
	case pbe.EventType_UPDATE, pbe.EventType_INSERT:
		msg.Body, msg.Columns = t.getBody(rowData.GetAfterColumns(), true)
		t.dmlHandler(msg)
	case pbe.EventType_DELETE:
		msg.Body, _ = t.getBody(rowData.GetBeforeColumns(), false)
		t.dmlHandler(msg)
	}
}

func (t *ReplicationSyncer) getBody(columns []*pbe.Column, updated bool) (res map[string]interface{}, updatedColumns []string) {
	res = make(map[string]interface{})
	updatedColumns = make([]string, 0)
	for _, col := range columns {
		if updated && col.Updated {
			updatedColumns = append(updatedColumns, col.GetName())
		}
		res[col.GetName()] = col.GetValue()
	}
	return
}

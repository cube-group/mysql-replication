# mysql-replication
mysql replication sync with canal-server

## k8s install canal-server
[kubernetes apply.yaml](k8s/apply.yaml)

## golang run demo
```go
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

package main

import (
	"github.com/cube-group/mysql-replication/core"
	"log"
)

func main() {
	syncer := core.NewReplicationSyncer(
		core.CanalOption{
			Host:        "192.168.4.154",
			Port:        30111,
			Destination: "test",
			Filter:      ".*\\..*", //all schema and all tables
			//Filter:"default\\..*",//database default and its all tables
			//Filter: "default\\.abc_.*",//database default and abc_ prefix tables
		},
		handleForDML,
	)
	// 无需担心offset，/home/admin/canal-server/conf/test/meta.dat和h2.mv.db记录了同步位置点
	log.Println("start: ", syncer.Start())
}

func handleForDML(msg core.ReplicationMessage) {
	log.Printf("[%v] Schema: %v Table: %v UpdatedColumns: %+v %+v", msg.EventType, msg.SchemaName, msg.TableName, msg.Columns, msg.Body)
}


```
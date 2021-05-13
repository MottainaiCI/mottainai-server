/*

Copyright (C) 2021 Daniele Rondina <geaaru@sabayonlinux.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/

package queues

import (
	"encoding/json"
	"reflect"
)

type NodeQueues struct {
	ID           string              `json:"ID"`
	AgentKey     string              `json:"akey" form:"akey"`
	NodeId       string              `json:"nodeid" form:"nodeid"`
	Queues       map[string][]string `json:"queues" form:"queues"`
	CreationDate string              `json:"creation_date"`
}

func NewNodeQueueFromJson(data []byte) NodeQueues {
	var q NodeQueues
	json.Unmarshal(data, &q)
	return q
}

func NewNodeQueuesFromMap(q map[string]interface{}) NodeQueues {
	var (
		nid  string
		akey string
		cd   string
	)

	queues := make(map[string][]string, 0)

	if str, ok := q["creation_date"].(string); ok {
		cd = str
	}

	if str, ok := q["nodeid"].(string); ok {
		nid = str
	}

	if str, ok := q["akey"].(string); ok {
		akey = str
	}

	if m, ok := q["queues"].(map[string]interface{}); ok {
		for k, v := range m {
			qt := []string{}
			tasks := v.([]interface{})

			if len(tasks) > 0 {
				for _, t := range tasks {
					qt = append(qt, t.(string))
				}
			}

			queues[k] = qt
		}
	}

	ans := NodeQueues{
		AgentKey:     akey,
		NodeId:       nid,
		Queues:       queues,
		CreationDate: cd,
	}

	return ans
}

func (t *NodeQueues) ToMap() map[string]interface{} {
	ts := make(map[string]interface{})
	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag
		ts[tag.Get("form")] = valueField.Interface()
	}

	return ts
}

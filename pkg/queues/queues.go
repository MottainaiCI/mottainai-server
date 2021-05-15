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

type Queue struct {
	ID           string   `json:"ID"`
	Qid          string   `form:"qid" json:"qid"`
	Name         string   `json:"name" form:"name"`
	Waiting      []string `json:"tasks_waiting" form:"waiting_tasks"`
	InProgress   []string `json:"tasks_inprogress" form:"tasks_inprogress"`
	CreationDate string   `json:"creation_date" form:"creation_date"`
	UpdateDate   string   `json:"update_date" form:"update_date"`
}

func NewFromJson(data []byte) Queue {
	var q Queue
	json.Unmarshal(data, &q)
	return q
}

func NewQueueFromMap(q map[string]interface{}) Queue {

	var (
		qid  string
		name string
		cd   string
		ud   string
	)

	progress := []string{}
	waiting := []string{}

	if str, ok := q["qid"].(string); ok {
		qid = str
	}

	if str, ok := q["name"].(string); ok {
		name = str
	}

	if str, ok := q["creation_date"].(string); ok {
		cd = str
	}

	if str, ok := q["update_date"].(string); ok {
		ud = str
	}

	if arr, ok := q["tasks_waiting"].([]interface{}); ok {
		if len(arr) > 0 {
			for _, t := range arr {
				waiting = append(waiting, t.(string))
			}
		}
	}

	if arr, ok := q["tasks_inprogress"].([]interface{}); ok {
		if len(arr) > 0 {
			for _, t := range arr {
				progress = append(progress, t.(string))
			}
		}
	}

	queue := Queue{
		Qid:          qid,
		Name:         name,
		Waiting:      waiting,
		InProgress:   progress,
		CreationDate: cd,
		UpdateDate:   ud,
	}
	return queue
}

func (t *Queue) ToMap() map[string]interface{} {

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

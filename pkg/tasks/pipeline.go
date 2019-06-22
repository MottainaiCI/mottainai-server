/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package agenttasks

import (
	"encoding/json"
	"reflect"
	"strconv"
	"time"

	"github.com/ghodss/yaml"

	"io/ioutil"
)

type Pipeline struct {
	ID string `json:"ID" form:"ID"` // ARMv7l overflows :(

	Chain []string        `json:"chain" form:"chain"`
	Chord []string        `json:"chord" form:"chord"`
	Group []string        `json:"group" form:"group"`
	Tasks map[string]Task `json:"tasks" form:"tasks"`

	Queue string `json:"queue" form:"queue"`
	//Status       string   `json:"status" form:"status"`
	//Result       string   `json:"result" form:"result"`
	Retry string `json:"retry" form:"retry"`

	Owner       string `json:"pipeline_owner_id" form:"pipeline_owner_id"`
	Name        string `json:"pipeline_name" form:"pipeline_name"`
	CreatedTime string `json:"created_time" form:"created_time"`
	StartTime   string `json:"start_time" form:"start_time"`
	EndTime     string `json:"end_time" form:"end_time"`
	Concurrency string `json:"concurrency" form:"concurrency"`
}

func PipelineFromJsonFile(file string) (*Pipeline, error) {
	var t *Pipeline
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal(content, &t); err != nil {
		return t, err
	}
	return t, nil
}

func PipelineFromYamlFile(file string) (*Pipeline, error) {
	var t *Pipeline
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return t, err
	}
	if err := yaml.Unmarshal(content, &t); err != nil {
		return t, err
	}
	return t, nil
}

func (t *Pipeline) Trials() int {

	ret, err := strconv.Atoi(t.Retry)
	if err != nil {
		return 0
	}

	return ret
}

func (t *Pipeline) Reset() {

	t.CreatedTime = time.Now().Format("20060102150405")
	t.EndTime = ""
	t.StartTime = ""
}

type PipelineForm struct {
	*Pipeline
	Tasks string
}

func (t *Pipeline) ToMap(serialize bool) map[string]interface{} {

	ts := make(map[string]interface{})
	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		// XXX: Otherwise gob is confused
		//fmt.Println(valueField.Type(), reflect.ValueOf(t.Tasks).Kind())
		if valueField.Kind() == reflect.ValueOf(t.Tasks).Kind() && serialize {
			m := make(map[string]interface{})
			elem, _ := valueField.Interface().(map[string]Task)

			for i, o := range elem {
				f := &o
				m[i] = f.ToMap()
			}

			ts[tag.Get("form")] = m
		} else {
			ts[tag.Get("form")] = valueField.Interface()
		}

		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("tag_name"))
	}
	return ts
}

// TODO: Port NewUserFromMap Task to same or make it common func
func NewPipelineFromMap(t map[string]interface{}) Pipeline {
	u := &Pipeline{}
	val := reflect.ValueOf(u).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag
		if typeField.Type.String() == "int" {
			if str, ok := t[tag.Get("form")].(int); ok {
				valueField.SetInt(int64(str))
			}
		}
		if typeField.Type.String() == "string" {
			if str, ok := t[tag.Get("form")].(string); ok {
				valueField.SetString(str)
			}
		}

		if typeField.Type.String() == "[]string" {
			if b, ok := t[tag.Get("form")].([]string); ok {
				valueField.Set(reflect.ValueOf(b))
			} else if b, ok := t[tag.Get("form")].([]interface{}); ok {
				// convert all to string before set
				var r []string
				for _, f := range b {
					r = append(r, f.(string))
				}
				valueField.Set(reflect.ValueOf(r))
			}
		}

		if typeField.Type.Kind() == reflect.ValueOf(u.Tasks).Kind() {
			if b, ok := t[tag.Get("form")].(map[string]interface{}); ok {

				m := make(map[string]Task)
				for k, v := range b {
					m[k] = NewTaskFromMap(v.(map[string]interface{}))
				}

				valueField.Set(reflect.ValueOf(m))
			} else if b, ok := t[tag.Get("form")].([]interface{}); ok {
				// convert all to string before set
				var r []string
				for _, f := range b {
					r = append(r, f.(string))
				}
				valueField.Set(reflect.ValueOf(r))
			}
		}
	}
	return *u
}

func NewPipelineFromJson(data []byte) Pipeline {
	var t Pipeline
	json.Unmarshal(data, &t)
	return t
}

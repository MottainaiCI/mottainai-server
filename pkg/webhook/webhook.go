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

package webhook

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"

	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"

	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"github.com/sethvargo/go-password/password"

	"reflect"
)

type WebHookSingle struct {
	WebHook  *WebHook
	Task     *task.Task
	Pipeline *task.Pipeline
}

type WebHook struct {
	ID       string `json:"id" form:"id"`
	Key      string `json:"key" form:"key"`
	Type     string `json:"type" form:"type"`
	URL      string `json:"url" form:"url"`
	OwnerId  string `json:"owner_id" form:"owner_id"`
	Task     string `json:"default_task" form:"default_task"`
	Pipeline string `json:"default_pipeline" form:"default_pipeline"`
}

func (t *WebHook) HasTask() bool {
	if len(t.Task) > 0 {
		return true
	}
	return false
}

func (t *WebHook) HasPipeline() bool {
	if len(t.Pipeline) > 0 {
		return true
	}
	return false
}

func (t *WebHook) SetPipeline(pipeline *task.Pipeline) error {
	str, err := utils.SerializeToString(pipeline)
	if err != nil {
		t.Pipeline = ""
		return err
	}
	t.Pipeline = str
	return nil
}

func (t *WebHook) ReadPipeline() (*task.Pipeline, error) {
	var pipeline *task.Pipeline
	buf, err := utils.DecodeString(t.Pipeline)
	if err != nil {
		return pipeline, err
	}
	d := gob.NewDecoder(buf)
	if err := d.Decode(&pipeline); err != nil {
		return nil, err
	}
	return pipeline, nil
}

func (t *WebHook) SetTask(ta *task.Task) error {
	str, err := utils.SerializeToString(ta)
	if err != nil {
		t.Task = ""
		return err
	}
	t.Task = str
	return nil
}

func (t *WebHook) ReadTask() (*task.Task, error) {
	var Task *task.Task
	buf, err := utils.DecodeString(t.Task)
	if err != nil {
		return Task, err
	}
	d := gob.NewDecoder(buf)
	if err := d.Decode(&Task); err != nil {
		return Task, err
	}
	return Task, nil
}

func GenerateUserWebHook(id string) (*WebHook, error) {
	t, err := GenerateWebHook()
	if err != nil {
		return t, err
	}
	t.OwnerId = id

	return t, nil
}

func Gen() (string, error) {
	res, err := password.Generate(64, 10, 10, false, false)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write([]byte(res))

	return hex.EncodeToString(h.Sum(nil)), nil
}

func GenerateWebHook() (*WebHook, error) {
	t := NewWebHook()
	res, err := Gen()
	if err != nil {
		return t, err
	}
	t.Key = res

	res, err = Gen()
	if err != nil {
		return t, err
	}
	t.URL = res

	return t, nil
}

func NewWebHook() *WebHook {
	return &WebHook{}
}

// TODO: Port NewWebHookFromMap Task to same or make it common func
func NewWebHookFromMap(t map[string]interface{}) WebHook {
	u := &WebHook{}
	val := reflect.ValueOf(u).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		if typeField.Type.Name() == "string" {
			if str, ok := t[tag.Get("form")].(string); ok {
				valueField.SetString(str)
			}
		}
		if typeField.Type.Name() == "int64" {

			if i, ok := t[tag.Get("form")].(int64); ok {
				valueField.SetInt(i)
			}
		}
		if typeField.Type.Name() == "int" {

			if i, ok := t[tag.Get("form")].(int); ok {
				valueField.SetInt(int64(i))
			}
		}
		if typeField.Type.Name() == "bool" {
			if b, ok := t[tag.Get("form")].(bool); ok {
				valueField.SetBool(b)
			}
		}
		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("tag_name"))
	}
	return *u
}

func NewWebHookFromJson(data []byte) WebHook {
	var t WebHook
	json.Unmarshal(data, &t)
	return t
}

func (t *WebHook) Clear() {
}

func (t *WebHook) ToMap() map[string]interface{} {

	ts := make(map[string]interface{})
	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		tag := typeField.Tag

		ts[tag.Get("form")] = valueField.Interface()
		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("tag_name"))
	}
	return ts
}

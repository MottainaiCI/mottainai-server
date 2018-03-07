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
	"os"
	"path"
	"reflect"
	"strconv"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type Task struct {
	ID           int    `json:"ID"`
	Source       string `json:"source" form:"source"`
	Script       string `json:"script" form:"script"`
	Yaml         string `json:"yaml" form:"yaml"`
	Directory    string `json:"directory" form:"directory"`
	TaskName     string `json:"task" form:"task"`
	Status       string `json:"status" form:"status"`
	Output       string `json:"output" form:"output"`
	Result       string `json:"result" form:"result"`
	Namespace    string `json:"namespace" form:"namespace"`
	Commit       string `json:"commit" form:"commit"`
	PrivKey      string `json:"privkey" form:"privkey"`
	AuthHosts    string `json:"authhosts" form:"authhosts"`
	Node         int    `json:"nodeid" form:"nodeid"`
	Owner        int    `json:"ownerid" form:"ownerid"`
	Image        string `json:"image" form:"image"`
	ExitStatus   string `json:"exit_status" form:"exit_status"`
	Storage      string `json:"storage" form:"storage"`
	ArtefactPath string `json:"artefact_path" form:"artefact_path"`
	StoragePath  string `json:"storage_path" form:"storage_path"`

	CreatedTime string `json:"created_time" form:"created_time"`
	StartTime   string `json:"start_time" form:"start_time"`
	EndTime     string `json:"end_time" form:"end_time"`
}

func (t *Task) ToMap() map[string]interface{} {

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

func (t *Task) IsRunning() bool {

	if t.Status == "running" {
		return true
	}
	return false
}
func (t *Task) IsWaiting() bool {

	if t.Status == "waiting" {
		return true
	}
	return false
}

func (t *Task) ClearBuildLog() {
	os.RemoveAll(path.Join(setting.Configuration.ArtefactPath, strconv.Itoa(t.ID), "build.log"))
}

func (t *Task) AppendBuildLog(s string) error {

	os.MkdirAll(path.Join(setting.Configuration.ArtefactPath, strconv.Itoa(t.ID)), 0777)
	file, err := os.OpenFile(path.Join(setting.Configuration.ArtefactPath, strconv.Itoa(t.ID), "build.log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err = file.WriteString(s + "\n"); err != nil {
		panic(err)
	}
	return nil

}

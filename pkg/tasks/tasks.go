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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/namespace"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	flock "github.com/theckman/go-flock"
)

type Task struct {
	ID           string   `json:"ID" form:"ID"` // ARMv7l overflows :(
	Source       string   `json:"source" form:"source"`
	Script       []string `json:"script" form:"script"`
	Directory    string   `json:"directory" form:"directory"`
	TaskName     string   `json:"task" form:"task"`
	Status       string   `json:"status" form:"status"`
	Output       string   `json:"output" form:"output"`
	Result       string   `json:"result" form:"result"`
	Entrypoint   []string `json:"entrypoint" form:"entrypoint"`
	Namespace    string   `json:"namespace" form:"namespace"`
	Commit       string   `json:"commit" form:"commit"`
	PrivKey      string   `json:"privkey" form:"privkey"`
	AuthHosts    string   `json:"authhosts" form:"authhosts"`
	Node         string   `json:"node_id" form:"node_id"`
	Owner        string   `json:"owner_id" form:"owner_id"`
	Image        string   `json:"image" form:"image"`
	ExitStatus   string   `json:"exit_status" form:"exit_status"`
	Storage      string   `json:"storage" form:"storage"`
	ArtefactPath string   `json:"artefact_path" form:"artefact_path"`
	StoragePath  string   `json:"storage_path" form:"storage_path"`
	RootTask     string   `json:"root_task" form:"root_task"`
	Prune        string   `json:"prune" form:"prune"`
	CacheImage   string   `json:"cache_image" form:"cache_image"`
	CacheClean   string   `json:"cache_clean" form:"cache_clean"`

	TagNamespace string `json:"tag_namespace" form:"tag_namespace"`

	CreatedTime string `json:"created_time" form:"created_time"`
	StartTime   string `json:"start_time" form:"start_time"`
	EndTime     string `json:"end_time" form:"end_time"`
	Queue       string `json:"queue" form:"queue"`

	Delayed     string   `json:"eta" form:"eta"`
	TimeOut     float64  `json:"timeout" form:"timeout"`
	Binds       []string `json:"binds" form:"binds"`
	Environment []string `json:"environment" form:"environment"`
}

type Plan struct {
	*Task
	Planned string `json:"planned" form:"planned"`
}

func (t *Plan) ToMap() map[string]interface{} {

	ts := make(map[string]interface{})
	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		ts[tag.Get("form")] = valueField.Interface()
	}
	val = reflect.ValueOf(t.Task).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		ts[tag.Get("form")] = valueField.Interface()
		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("tag_name"))
	}
	return ts
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

func FromFile(file string) (*Task, error) {
	var t *Task
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal(content, &t); err != nil {
		return t, err
	}
	return t, nil
}

func (t *Task) Reset() {
	t.Output = ""
	t.Result = setting.TASK_RESULT_UNKNOWN
	t.Status = ""
	t.ExitStatus = ""
	t.CreatedTime = time.Now().Format("20060102150405")
	t.EndTime = ""
	t.Owner = ""
	t.Node = ""
	t.StartTime = ""
}

func (t *Task) IsOwner(id int) bool {

	if strconv.Itoa(id) == t.Owner {
		return true
	}

	return false
}

func (t *Task) IsRunning() bool {

	if t.Status == setting.TASK_STATE_RUNNING {
		return true
	}
	return false
}
func (t *Task) IsWaiting() bool {

	if t.Status == setting.TASK_STATE_WAIT {
		return true
	}
	return false
}

func (t *Task) ClearBuildLog() {
	os.RemoveAll(path.Join(setting.Configuration.ArtefactPath, t.ID, "build_"+t.ID+".log"))
}

func (t *Task) Clear() {
	os.RemoveAll(path.Join(setting.Configuration.ArtefactPath, t.ID))
	os.RemoveAll(path.Join(setting.Configuration.LockPath, t.ID+".lock"))
}

func (t *Task) GetLogPart(pos int) string {
	var b3 []byte
	err := t.LockSection(func() error {
		file, err := os.Open(path.Join(setting.Configuration.ArtefactPath, t.ID, "build_"+t.ID+".log"))
		if err != nil {
			return err
		}
		_, err = file.Seek(int64(pos), 0)
		if err != nil {
			return err
		}
		fi, err := file.Stat()
		if err != nil {
			return err
		}

		b3 = make([]byte, fi.Size()-int64(pos))
		_, err = file.Read(b3)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return ""
	}
	return string(b3)
}

func (t *Task) TailLog(pos int) string {
	var b3 []byte
	err := t.LockSection(func() error {
		file, err := os.Open(path.Join(setting.Configuration.ArtefactPath, t.ID, "build_"+t.ID+".log"))
		if err != nil {
			return err
		}

		fi, err := file.Stat()
		if err != nil {
			return err
		}

		where := fi.Size() - int64(pos)
		if int64(pos) > fi.Size() {
			where = 0
		}

		_, err = file.Seek(where, 0)
		if err != nil {
			return err
		}

		b3 = make([]byte, int64(pos))
		_, err = file.Read(b3)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return ""
	}
	return string(b3)
}

func (t *Task) LockSection(f func() error) error {
	os.MkdirAll(setting.Configuration.LockPath, os.ModePerm)
	lockfile := path.Join(setting.Configuration.LockPath, t.ID+".lock")
	fileLock := flock.NewFlock(lockfile)

	locked, err := fileLock.TryLock()
	if err != nil {
		return err
	}

	if locked {
		err := f()
		fileLock.Unlock()
		return err
	}
	return nil
}

func (t *Task) AppendBuildLog(s string) error {

	os.MkdirAll(path.Join(setting.Configuration.ArtefactPath, t.ID), os.ModePerm)
	return t.LockSection(func() error {

		file, err := os.OpenFile(path.Join(setting.Configuration.ArtefactPath, t.ID, "build_"+t.ID+".log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}

		defer file.Close()

		if _, err = file.WriteString(s + "\n"); err != nil {
			return err
		}
		return nil
	})

}

func (t *Task) IsStopped() bool {
	if t.Status == setting.TASK_STATE_STOPPED || t.Status == setting.TASK_STATE_ASK_STOP {
		return true
	}

	return false
}
func (t *Task) IsDone() bool {
	if t.Status == setting.TASK_STATE_DONE {
		return true
	}

	return false
}

func (t *Task) WantsClean() bool {
	if len(t.CacheClean) > 0 {
		return true
	}

	return false
}

func (t *Task) IsSuccess() bool {
	if t.ExitStatus == "0" {
		return true
	}

	return false
}

func (t *Task) HandleStatus() {
	fmt.Println("Handlestatus called")
	if t.Status == setting.TASK_STATE_DONE {
		if t.ExitStatus == "0" {
			t.OnSuccess()
		} else {
			t.OnFailure()
		}
		t.Done()
	}
}

func (t *Task) DecodeStatus(state string) string {
	if state == "0" {
		return setting.TASK_RESULT_SUCCESS
	}

	return setting.TASK_RESULT_FAILED
}

func (t *Task) Artefacts() []string {
	return utils.TreeList(filepath.Join(setting.Configuration.ArtefactPath, t.ID))
}

func (t *Task) Done() {
	fmt.Println("Build done")
}

func (t *Task) OnFailure() {
	fmt.Println("Build failed")

}

func (t *Task) OnSuccess() {
	fmt.Println("Build succeeded")

	if len(t.TagNamespace) > 0 {
		ns := namespace.NewFromMap(map[string]interface{}{"name": t.TagNamespace, "path": t.TagNamespace})
		ns.Tag(t.ID)
	}
}

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
	"errors"
	"strconv"

	"github.com/MottainaiCI/mottainai-server/pkg/client"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/backends"
	machinerytask "github.com/RichardKnop/machinery/v1/tasks"
)

type Task struct {
	ID         int    `json:"ID"`
	Source     string `json:"source" form:"source"`
	Script     string `json:"script" form:"script"`
	Yaml       string `json:"yaml" form:"yaml"`
	Directory  string `json:"directory" form:"directory"`
	TaskName   string `json:"task" form:"task"`
	Status     string `json:"status" form:"status"`
	Output     string `json:"output" form:"output"`
	Result     string `json:"result" form:"result"`
	Namespace  string `json:"namespace" form:"namespace"`
	Commit     string `json:"commit" form:"commit"`
	PrivKey    string `json:"privkey" form:"privkey"`
	AuthHosts  string `json:"authhosts" form:"authhosts"`
	Node       int    `json:"nodeid" form:"nodeid"`
	Owner      int    `json:"ownerid" form:"ownerid"`
	Image      string `json:"image" form:"image"`
	ExitStatus string `json:"exit_status" form:"exit_status"`
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

var AvailableTasks = map[string]interface{}{
	"execute": Execute,
	"error":   HandleErr,
	"success": HandleSuccess,
}

func NewTaskFromJson(data []byte) Task {
	var t Task
	json.Unmarshal(data, &t)
	return t
}

func NewTaskFromMap(t map[string]interface{}) Task {

	var (
		source      string
		script      string
		yaml        string
		directory   string
		namespace   string
		commit      string
		taskname    string
		output      string
		image       string
		status      string
		result      string
		exit_status string
	)
	if str, ok := t["exit_status"].(string); ok {
		exit_status = str
	}
	if str, ok := t["source"].(string); ok {
		source = str
	}
	if str, ok := t["script"].(string); ok {
		script = str
	}
	if str, ok := t["yaml"].(string); ok {
		yaml = str
	}
	if str, ok := t["directory"].(string); ok {
		directory = str
	}
	if str, ok := t["task"].(string); ok {
		taskname = str
	}
	if str, ok := t["namespace"].(string); ok {
		namespace = str
	}
	if str, ok := t["commit"].(string); ok {
		commit = str
	}
	if str, ok := t["output"].(string); ok {
		output = str
	}
	if str, ok := t["result"].(string); ok {
		result = str
	}
	if str, ok := t["status"].(string); ok {
		status = str
	}
	if _, ok := AvailableTasks[taskname]; !ok {
		taskname = ""
	}
	if str, ok := t["image"].(string); ok {
		image = str
	}
	task := Task{
		Source:     source,
		Script:     script,
		Yaml:       yaml,
		Directory:  directory,
		TaskName:   taskname,
		Namespace:  namespace,
		Commit:     commit,
		Output:     output,
		Result:     result,
		Status:     status,
		Image:      image,
		ExitStatus: exit_status,
	}
	return task
}

func SendTask(rabbit *machinery.Server, taskname string, taskid int) (*backends.AsyncResult, error) {

	if len(taskname) == 0 {
		return &backends.AsyncResult{}, errors.New("No task name specified")
	}
	if _, ok := AvailableTasks[taskname]; !ok {
		return &backends.AsyncResult{}, errors.New("No task name specified")
	}
	onErr := make([]*machinerytask.Signature, 0)

	onErr = append(onErr, &machinerytask.Signature{
		Name: "error",
		Args: []machinerytask.Arg{
			{
				Type:  "string",
				Value: strconv.Itoa(taskid),
			},
		},
	})

	onSuccess := make([]*machinerytask.Signature, 0)

	onSuccess = append(onSuccess, &machinerytask.Signature{
		Name: "success",
		Args: []machinerytask.Arg{
			{
				Type:  "string",
				Value: strconv.Itoa(taskid),
			},
		},
	})

	return rabbit.SendTask(&machinerytask.Signature{
		Name: taskname,
		Args: []machinerytask.Arg{
			{
				Type:  "string",
				Value: strconv.Itoa(taskid),
			},
		},

		OnError:   onErr,
		OnSuccess: onSuccess,
	})
}

func RegisterTasks(m *machinery.Server) {
	err := m.RegisterTasks(AvailableTasks)
	if err != nil {
		panic(err)
	}
}

func FetchTask(fetcher *client.Fetcher, docID string) Task {
	task_data, err := fetcher.GetTask()

	if err != nil {
		panic(err)
	}
	return NewTaskFromJson(task_data)
}

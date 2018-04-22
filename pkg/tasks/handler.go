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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/backends"
	machinerytask "github.com/RichardKnop/machinery/v1/tasks"
)

type TaskHandler struct {
	Tasks map[string]interface{}
}

func (h *TaskHandler) Exists(s string) bool {
	if _, ok := h.Tasks[s]; ok {
		return true
	}
	return false
}

func (h *TaskHandler) Handler(s string) func(string) (int, error) {
	if f, ok := h.Tasks[s]; ok {
		return f.(func(string) (int, error))
	}
	panic(errors.New("No task handler found!"))
}

func DefaultTaskHandler() *TaskHandler {
	return &TaskHandler{Tasks: map[string]interface{}{
		"docker_execute": DockerPlayer,
		"error":          HandleErr,
		"success":        HandleSuccess,
	}}
}

func DockerPlayer(docID string) (int, error) {
	player := NewPlayer(docID)
	return player.Start(&DockerExecutor{})
}

func (h *TaskHandler) NewPlanFromJson(data []byte) Plan {
	var t Plan
	json.Unmarshal(data, &t)
	return t
}

func (h *TaskHandler) NewTaskFromJson(data []byte) Task {
	var t Task
	json.Unmarshal(data, &t)
	return t
}

func (h *TaskHandler) NewPlanFromMap(t map[string]interface{}) Plan {
	tk := h.NewTaskFromMap(t)
	var planned string
	if str, ok := t["planned"].(string); ok {
		planned = str
	}
	pl := Plan{Task: &tk}
	pl.Planned = planned
	return pl
}

func (h *TaskHandler) NewTaskFromMap(t map[string]interface{}) Task {

	var (
		source        string
		script        string
		directory     string
		namespace     string
		commit        string
		taskname      string
		output        string
		image         string
		status        string
		result        string
		exit_status   string
		created_time  string
		start_time    string
		end_time      string
		storage       string
		storage_path  string
		artefact_path string
		root_task     string
		prune         string
		tag_namespace string
		cache_image   string

		environment []string
		binds       []string
	)
	if arr, ok := t["binds"].([]interface{}); ok {
		for _, v := range arr {
			binds = append(binds, v.(string))
		}
	}

	if arr, ok := t["environment"].([]interface{}); ok {
		for _, v := range arr {
			environment = append(environment, v.(string))
		}
	}

	if str, ok := t["root_task"].(string); ok {
		root_task = str
	}
	if str, ok := t["exit_status"].(string); ok {
		exit_status = str
	}
	if str, ok := t["source"].(string); ok {
		source = str
	}
	if str, ok := t["script"].(string); ok {
		script = str
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

	if str, ok := t["tag_namespace"].(string); ok {
		tag_namespace = str
	}
	if str, ok := t["status"].(string); ok {
		status = str
	}
	if !h.Exists(taskname) {
		taskname = ""
	}
	if str, ok := t["image"].(string); ok {
		image = str
	}
	if str, ok := t["storage"].(string); ok {
		storage = str
	}
	if str, ok := t["created_time"].(string); ok {
		created_time = str
	}
	if str, ok := t["start_time"].(string); ok {
		start_time = str
	}
	if str, ok := t["end_time"].(string); ok {
		end_time = str
	}
	if str, ok := t["storage_path"].(string); ok {
		storage_path = str
	}
	if str, ok := t["artefact_path"].(string); ok {
		artefact_path = str
	}
	if str, ok := t["prune"].(string); ok {
		prune = str
	}

	if str, ok := t["cache_image"].(string); ok {
		cache_image = str
	}

	task := Task{
		Source:       source,
		Script:       script,
		Directory:    directory,
		TaskName:     taskname,
		Namespace:    namespace,
		Commit:       commit,
		Output:       output,
		Result:       result,
		Status:       status,
		Storage:      storage,
		StoragePath:  storage_path,
		ArtefactPath: artefact_path,
		Image:        image,
		ExitStatus:   exit_status,
		CreatedTime:  created_time,
		StartTime:    start_time,
		EndTime:      end_time,
		RootTask:     root_task,
		TagNamespace: tag_namespace,
		Prune:        prune,
		CacheImage:   cache_image,
		Environment:  environment,
		Binds:        binds,
	}
	return task
}

func (h *TaskHandler) SendTask(rabbit *machinery.Server, taskname string, taskid int) (*backends.AsyncResult, error) {

	if !h.Exists(taskname) {
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

func (h *TaskHandler) RegisterTasks(m *machinery.Server) {
	th := DefaultTaskHandler()
	err := m.RegisterTasks(th.Tasks)
	if err != nil {
		panic(err)
	}
}

func (h *TaskHandler) FetchTask(fetcher *client.Fetcher) Task {
	task_data, err := fetcher.GetTask()

	if err != nil {
		panic(err)
	}
	return h.NewTaskFromJson(task_data)
}

func (h *TaskHandler) UploadArtefact(fetcher *client.Fetcher, path, art string) error {

	_, file := filepath.Split(path)
	rel := strings.Replace(path, art, "", 1)
	rel = strings.Replace(rel, file, "", 1)

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		// do directory stuff
		return err
	case mode.IsRegular():
		fetcher.AppendTaskOutput("Uploading " + path + " to " + rel)
		fetcher.UploadArtefact(path, rel)
	}

	return nil
}

func HandleSuccess(docID string, result int) error {
	fetcher := client.NewFetcher(docID)

	fetcher.SetTaskField("exit_status", strconv.Itoa(result))
	if result != 0 {
		fetcher.SetTaskResult("failed")
	} else {
		fetcher.SetTaskResult("success")
	}

	th := DefaultTaskHandler()

	task_info := th.FetchTask(fetcher)
	if task_info.Status != "stop" {
		fetcher.SetTaskStatus("done")
	} else {
		fetcher.SetTaskStatus("stopped")
	}
	return nil
}

func HandleErr(errstring, docID string) error {
	fetcher := client.NewFetcher(docID)

	fetcher.AppendTaskOutput(errstring)
	fetcher.SetTaskResult("error")

	th := DefaultTaskHandler()

	task_info := th.FetchTask(fetcher)
	if task_info.Status != "stop" {
		fetcher.SetTaskStatus("done")
	} else {
		fetcher.SetTaskStatus("stopped")
	}
	return nil
}

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

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	machinery "github.com/RichardKnop/machinery/v1"
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
		//	"success":        HandleSuccess,
	}}
}

func HandleArgs(args ...interface{}) (string, int, error) {
	var docID string
	if len(args) > 1 {
		docID = args[0].(string)

		for _, v := range args[1:] { // If other tasks in the chain failed, propagate same error
			if v.(int) != 0 {
				return docID, v.(int), errors.New("Other tasks in the chain failed!")
			}
		}
	} else {
		docID = args[len(args)-1].(string)
	}
	return docID, 0, nil
}

func DockerPlayer(args ...interface{}) (int, error) {
	docID, e, err := HandleArgs(args...)
	player := NewPlayer(docID)
	executor := NewDockerExecutor()
	executor.MottainaiClient = client.NewTokenClient(setting.Configuration.AppURL, setting.Configuration.ApiKey)
	if err != nil {
		player.EarlyFail(executor, docID, err.Error())
		return e, err
	}

	return player.Start(executor)
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
		script        []string
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
		cache_clean   string
		queue         string
		owner, node   string
		environment   []string
		binds         []string
	)

	binds = make([]string, 0)
	environment = make([]string, 0)
	script = make([]string, 0)

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

	if arr, ok := t["script"].([]interface{}); ok {
		for _, v := range arr {
			script = append(script, v.(string))
		}
	}
	if i, ok := t["owner_id"].(string); ok {
		owner = i
	}
	if i, ok := t["node_id"].(string); ok {
		node = i
	}
	if str, ok := t["queue"].(string); ok {
		queue = str
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
	if str, ok := t["cache_clean"].(string); ok {
		cache_clean = str
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

	var timeout float64
	if str, ok := t["timeout"].(float64); ok {
		timeout = str
	}
	var delayed string
	if str, ok := t["string"].(string); ok {
		delayed = str
	}
	var entrypoint []string
	entrypoint = make([]string, 0)

	if arr, ok := t["entrypoint"].([]interface{}); ok {
		for _, v := range arr {
			entrypoint = append(entrypoint, v.(string))
		}
	}
	var id string

	if str, ok := t["ID"].(string); ok {
		id = str
	}

	task := Task{
		ID:           id,
		Queue:        queue,
		Source:       source,
		Script:       script,
		Delayed:      delayed,
		Directory:    directory,
		TaskName:     taskname,
		Namespace:    namespace,
		Commit:       commit,
		Entrypoint:   entrypoint,
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
		Node:         node,
		Prune:        prune,
		CacheImage:   cache_image,
		Environment:  environment,
		Binds:        binds,
		CacheClean:   cache_clean,
		Owner:        owner,
		TimeOut:      timeout,
	}
	return task
}

func (h *TaskHandler) RegisterTasks(m *machinery.Server) {
	th := DefaultTaskHandler()
	err := m.RegisterTasks(th.Tasks)
	if err != nil {
		panic(err)
	}
}

func (h *TaskHandler) FetchTask(fetcher client.HttpClient) Task {
	task_data, err := fetcher.GetTask()

	if err != nil {
		panic(err)
	}
	return h.NewTaskFromJson(task_data)
}

func HandleSuccess(docID string, result int) error {
	fetcher := client.NewFetcher(docID)
	fetcher.Token = setting.Configuration.ApiKey
	res := strconv.Itoa(result)
	fetcher.SetTaskField("exit_status", res)
	if result != 0 {
		fetcher.FailTask("Exited with " + res)
	} else {
		fetcher.SuccessTask()
	}

	th := DefaultTaskHandler()

	task_info := th.FetchTask(fetcher)
	if task_info.Status != setting.TASK_STATE_ASK_STOP {
		fetcher.FinishTask()
	} else {
		fetcher.AbortTask()
	}
	return nil
}

func HandleErr(errstring, docID string) error {
	fetcher := client.NewFetcher(docID)
	fetcher.Token = setting.Configuration.ApiKey

	fetcher.AppendTaskOutput(errstring)

	th := DefaultTaskHandler()

	task_info := th.FetchTask(fetcher)
	if task_info.Status != setting.TASK_STATE_ASK_STOP {
		fetcher.FinishTask()
	} else {
		fetcher.AbortTask()
	}

	fetcher.ErrorTask()
	return nil
}

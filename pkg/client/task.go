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

package client

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"

	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
)

func (f *Fetcher) TaskLog(id string) ([]byte, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("stream_output"),
		Options: map[string]interface{}{
			":id":  id,
			":pos": "0",
		},
	}

	var res []byte
	var err error

	f.HandleRaw(req, func(b io.ReadCloser) error {
		res, err = ioutil.ReadAll(b)
		return err
	})
	return res, err
}

func (f *Fetcher) TaskDelete(id string) (event.APIResponse, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("delete"),
		Options: map[string]interface{}{
			":id": id,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) SetTaskField(field, value string) (event.APIResponse, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("update_field"),
		Options: map[string]interface{}{
			"id":    f.docID,
			"field": field,
			"value": value,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) SetTaskStatus(status string) (event.APIResponse, error) {

	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("update"),
		Options: map[string]interface{}{
			"id":     f.docID,
			"status": status,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) AbortTask() {
	f.SetTaskResult("")
	f.SetTaskStatus(setting.TASK_STATE_STOPPED)
}

func (f *Fetcher) FailTask(e string) {
	f.SetTaskResult(setting.TASK_RESULT_FAILED)
	f.AppendTaskOutput(e)
	f.FinishTask()
}

func (f *Fetcher) StartTask(id string) (event.APIResponse, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("start"),
		Options: map[string]interface{}{
			":id": id,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) StopTask(id string) (event.APIResponse, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("stop"),
		Options: map[string]interface{}{
			":id": id,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) CreateTask(taskdata map[string]interface{}) (event.APIResponse, error) {
	req := schema.Request{
		Route:   v1.Schema.GetTaskRoute("create"),
		Options: taskdata,
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) CloneTask(id string) (event.APIResponse, error) {

	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("clone"),
		Options: map[string]interface{}{
			":id": id,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) SetupTask() (event.APIResponse, error) {

	f.SetTaskStatus(setting.TASK_STATE_SETUP)

	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("update_node"),
		Options: map[string]interface{}{
			"id":  f.docID,
			"key": f.Config.GetAgent().AgentKey,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) RunTask() {
	f.SetTaskStatus(setting.TASK_STATE_RUNNING)
}

func (f *Fetcher) ErrorTask() {
	f.SetTaskResult(setting.TASK_RESULT_ERROR)
}

func (f *Fetcher) FinishTask() {
	f.SetTaskStatus(setting.TASK_STATE_DONE)
}

func (f *Fetcher) SuccessTask() {
	f.SetTaskResult(setting.TASK_RESULT_SUCCESS)
	f.FinishTask()
}

func (f *Fetcher) GetTask() ([]byte, error) {

	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("as_json"),
		Options: map[string]interface{}{
			":id": f.docID,
		},
	}

	var res []byte
	var err error

	f.HandleRaw(req, func(b io.ReadCloser) error {
		res, err = ioutil.ReadAll(b)
		return err
	})
	return res, err
}

func (f *Fetcher) TaskLogArtefact(id string) ([]byte, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("task_log"),
		Options: map[string]interface{}{
			":id": id,
		},
	}

	var res []byte
	var err error

	f.HandleRaw(req, func(b io.ReadCloser) error {
		res, err = ioutil.ReadAll(b)
		return err
	})
	return res, err
}

func (f *Fetcher) TaskStream(id, pos string) ([]byte, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("stream_output"),
		Options: map[string]interface{}{
			":id":  id,
			":pos": pos,
		},
	}

	var res []byte
	var err error

	f.HandleRaw(req, func(b io.ReadCloser) error {
		res, err = ioutil.ReadAll(b)
		return err
	})
	return res, err
}

func (f *Fetcher) AllTasks() ([]byte, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("show_all"),
	}

	var res []byte
	var err error

	f.HandleRaw(req, func(b io.ReadCloser) error {
		res, err = ioutil.ReadAll(b)
		return err
	})
	return res, err
}

func (f *Fetcher) SetTaskResult(result string) (event.APIResponse, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("update"),
		Options: map[string]interface{}{
			"id":     f.docID,
			"result": result,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) SetTaskOutput(output string) (event.APIResponse, error) {
	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("update"),
		Options: map[string]interface{}{
			"id":     f.docID,
			"output": output,
		},
	}

	return f.HandleAPIResponse(req)
}

func (f *Fetcher) StreamOutput(r io.Reader) {
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			f.AppendTaskOutput(scanner.Text() + "\n")
		}
		if err := scanner.Err(); err != nil {
			f.AppendTaskOutput("There was an error with the scanner in attached container " + err.Error() + "\n")
		}
	}(r)
}

func (f *Fetcher) AppendTaskOutput(output string) (event.APIResponse, error) {
	if f.ActiveReports {
		fmt.Println(output)
	}

	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("append"),
		Options: map[string]interface{}{
			"id":     f.docID,
			"output": output,
		},
	}

	return f.HandleAPIResponse(req)
}

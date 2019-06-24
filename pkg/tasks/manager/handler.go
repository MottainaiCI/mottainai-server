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
	"errors"
	"strconv"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	executors "github.com/MottainaiCI/mottainai-server/pkg/tasks/executors"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	machinery "github.com/RichardKnop/machinery/v1"
)

type TaskHandler struct {
	Tasks  map[string]interface{}
	Config *setting.Config
	Err    error
}
type Handler func(string) (int, error)

func (h *TaskHandler) AddHandler(s string, handler Handler) {
	h.Tasks[s] = handler
}
func (h *TaskHandler) RemoveHandler(s string) {
	delete(h.Tasks, s)
}
func (h *TaskHandler) Exists(s string) bool {
	if _, ok := h.Tasks[s]; ok {
		return true
	}
	return false
}

func (h *TaskHandler) Handler(s string) Handler {
	if f, ok := h.Tasks[s]; ok {
		return f.(Handler)
	}
	panic(errors.New("No task handler found!"))
}

var singletonTaskHandler *TaskHandler

func SetSingleton(th *TaskHandler) {
	singletonTaskHandler = th
}

func DefaultTaskHandler(config *setting.Config) *TaskHandler {
	if singletonTaskHandler != nil {
		return singletonTaskHandler
	}
	th := GenDefaultTaskHandler(config)
	SetSingleton(th)
	return th
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

func DockerPlayer(config *setting.Config) func(args ...interface{}) (int, error) {
	return func(args ...interface{}) (int, error) {
		docID, e, err := HandleArgs(args...)
		player := NewPlayer(docID)
		executor := executors.NewDockerExecutor(config)
		executor.MottainaiClient = client.NewTokenClient(
			config.GetWeb().AppURL,
			config.GetAgent().ApiKey, config)
		if err != nil {
			player.EarlyFail(executor, docID, err.Error())
			return e, err
		}

		return player.Start(executor)
	}
}

func KubernetesPlayer(config *setting.Config) func(args ...interface{}) (int, error) {
	return func(args ...interface{}) (int, error) {
		docID, e, err := HandleArgs(args...)
		player := NewPlayer(docID)
		executor := executors.NewKubernetesExecutor(config)
		executor.MottainaiClient = client.NewTokenClient(
			config.GetWeb().AppURL,
			config.GetAgent().ApiKey, config)
		if err != nil {
			player.EarlyFail(executor, docID, err.Error())
			return e, err
		}

		return player.Start(executor)
	}
}

func LibvirtPlayer(config *setting.Config) func(args ...interface{}) (int, error) {
	return func(args ...interface{}) (int, error) {
		docID, e, err := HandleArgs(args...)
		player := NewPlayer(docID)
		executor := executors.NewVagrantExecutor(config)
		executor.Provider = "libvirt"
		executor.MottainaiClient = client.NewTokenClient(
			config.GetWeb().AppURL,
			config.GetAgent().ApiKey, config)
		if err != nil {
			player.EarlyFail(executor, docID, err.Error())
			return e, err
		}

		return player.Start(executor)
	}
}

func VirtualBoxPlayer(config *setting.Config) func(args ...interface{}) (int, error) {
	return func(args ...interface{}) (int, error) {
		docID, e, err := HandleArgs(args...)
		player := NewPlayer(docID)
		executor := executors.NewVagrantExecutor(config)
		executor.Provider = "virtualbox"
		executor.MottainaiClient = client.NewTokenClient(
			config.GetWeb().AppURL,
			config.GetAgent().ApiKey, config)
		if err != nil {
			player.EarlyFail(executor, docID, err.Error())
			return e, err
		}

		return player.Start(executor)
	}
}

func (h *TaskHandler) RegisterTasks(m *machinery.Server) {
	th := DefaultTaskHandler(h.Config)
	err := m.RegisterTasks(th.Tasks)
	if err != nil {
		panic(err)
	}
}

func (h *TaskHandler) FetchTask(fetcher client.HttpClient) tasks.Task {
	t, err := tasks.FetchTask(fetcher)

	if err != nil {
		h.Err = err
	}
	return t
}

func HandleSuccess(config *setting.Config) func(docID string, result int) error {
	return func(docID string, result int) error {
		fetcher := client.NewFetcher(docID, config)
		fetcher.SetToken(config.GetAgent().ApiKey)
		res := strconv.Itoa(result)
		fetcher.SetTaskField("exit_status", res)
		if result != 0 {
			fetcher.FailTask("Exited with " + res)
		} else {
			fetcher.SuccessTask()
		}

		th := DefaultTaskHandler(config)

		task_info := th.FetchTask(fetcher)
		if task_info.Status != setting.TASK_STATE_ASK_STOP {
			fetcher.FinishTask()
		} else {
			fetcher.AbortTask()
		}
		return nil
	}
}

func HandleErr(config *setting.Config) func(errstring, docID string) error {
	return func(errstring, docID string) error {
		fetcher := client.NewFetcher(docID, config)
		fetcher.SetToken(config.GetAgent().ApiKey)

		fetcher.AppendTaskOutput(errstring)

		th := DefaultTaskHandler(config)

		task_info := th.FetchTask(fetcher)
		if task_info.Status != setting.TASK_STATE_ASK_STOP {
			fetcher.FinishTask()
		} else {
			fetcher.AbortTask()
		}

		fetcher.ErrorTask()
		return nil
	}
}

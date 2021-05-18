/*

Copyright (C) 2021  Daniele Rondina <geaaru@sabayonlinux.org>
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
	"fmt"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	executors "github.com/MottainaiCI/mottainai-server/pkg/tasks/executors"
)

const (
	// Number of seconds a task remain in Expiring state
	TASK_RETENTION_PERIOD = 60 * 10
)

type TaskManager struct {
	NodeId        string
	NodeUniqueId  string
	Fetcher       client.HttpClient
	Players       *TaskHandler
	RunningTasks  map[string]int64
	ExpiringTasks map[string]int64
	ClosingPhase  bool
}

func NewTaskManager(config *setting.Config) *TaskManager {
	ans := &TaskManager{
		Players: GenDefaultTaskHandler(config),
		Fetcher: client.NewTokenClient(
			config.GetWeb().AppURL,
			config.GetAgent().ApiKey,
			config),
		RunningTasks:  make(map[string]int64, 0),
		ExpiringTasks: make(map[string]int64, 0),
		ClosingPhase:  false,
	}

	return ans
}

func (tm *TaskManager) GetTasks() error {
	nodeQueue, err := tm.Fetcher.NodeQueueGetTasks(tm.NodeUniqueId)
	if err != nil {
		return err
	}

	return tm.AnalyzeQueues(nodeQueue.Queues)
}

func (tm *TaskManager) AnalyzeQueues(queues map[string][]string) error {
	mtasks := make(map[string]int, 0)

	// If the queues are empty we cleanup all running tasks
	if len(queues) == 0 {
		if len(tm.RunningTasks) > 0 {
			for tid, _ := range tm.RunningTasks {
				// POST: I consider tasks completed
				tm.ExpiringTasks[tid] = time.Now().UTC().Unix()
				fmt.Println(fmt.Sprintf("Move task %s to expiring state", tid))
			}

			tm.RunningTasks = make(map[string]int64, 0)
		}

	} else {

		for _, tasks := range queues {
			for _, t := range tasks {
				mtasks[t] = 1
				err := tm.HandleTask(t)
				if err != nil {
					fmt.Println("ERROR ON PROCESSING TASK ", t)
				}
			}
		}
	}

	// Check for completed tasks in Running tasks
	if len(tm.RunningTasks) > 0 {
		completedTasks := []string{}
		for tid, _ := range tm.RunningTasks {
			if _, ok := mtasks[tid]; !ok {
				tm.ExpiringTasks[tid] = time.Now().UTC().Unix()
				fmt.Println(fmt.Sprintf("Move task %s to expiring state", tid))
				completedTasks = append(completedTasks, tid)
			}
		}

		if len(completedTasks) > 0 {
			for _, tid := range completedTasks {
				delete(tm.RunningTasks, tid)
			}
		}
	}

	// Check for expired tasks
	if len(tm.ExpiringTasks) > 0 {

		nowTime := time.Now().UTC().Unix()
		expiredTasks := []string{}

		for tid, changeDate := range tm.ExpiringTasks {
			retentionTime := nowTime - changeDate

			if retentionTime > TASK_RETENTION_PERIOD {
				expiredTasks = append(expiredTasks, tid)
			}
		}

		if len(expiredTasks) > 0 {
			for _, tid := range expiredTasks {
				fmt.Println(fmt.Sprintf("Task %s is expired", tid))
				delete(tm.ExpiringTasks, tid)
			}
		}
	}

	return nil
}

func (tm *TaskManager) HandleTask(tid string) error {
	// Check if the task is already running
	if _, ok := tm.RunningTasks[tid]; ok {
		fmt.Println("Task is already running. Nothing to do.")
		return nil
	}

	// Check if the task is already been executed
	if _, ok := tm.ExpiringTasks[tid]; ok {
		fmt.Println("Task already executed. Nothing to do.")
		return nil
	}

	// I hate execute two time this call but it's temporary.
	// Handler and executors need refactor.
	task_info, err := tasks.FetchTask(tm.Fetcher, tid)
	if err != nil {
		return err
	}

	if !tm.Players.Exists(task_info.Type) {
		msg := "Unexpected task related to type " + task_info.Type + " not supported."
		fmt.Println(msg)
		// TODO: probably se set the task in failure
		return errors.New(msg)
	}

	tm.RunningTasks[tid] = time.Now().UTC().Unix()

	go tm.RunPlayer(task_info)

	return nil
}

func (tm *TaskManager) RunPlayer(task_info tasks.Task) error {
	var fn func(string, string) (*Player, executors.Executor)

	fn = tm.Players.Handler(task_info.Type)

	// TODO: handle this with a better way and WaitGroup
	player, executor := fn(task_info.ID, tm.NodeUniqueId)

	res, err := player.Start(executor)

	// TODO: handle retry
	if err != nil {
		tm.Players.HandleErr(err.Error(), task_info.ID)
	} else {
		tm.Players.HandleSuccess(task_info.ID, res)
	}

	// I run close here because on retries i will call
	// Close() only one time.
	executor.Close()

	// TODO: handle error

	_, err_del := tm.Fetcher.NodeQueueDelTask(
		tm.Players.Config.GetAgent().AgentKey,
		tm.NodeId,
		task_info.Queue,
		task_info.ID,
	)
	if err_del != nil {
		fmt.Println(fmt.Sprintf("Error on delete task %s from queue: %s",
			task_info.ID, err_del.Error()))
	}

	return err
}
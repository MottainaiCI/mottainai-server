/*
Copyright (C) 2021 Daniele Rondina <geaaru@sabayonlinux.org>

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

package scheduler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	setting "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/config"
	specs "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/specs"
	"github.com/MottainaiCI/mottainai-server/pkg/client"
	"github.com/MottainaiCI/mottainai-server/pkg/nodes"
	"github.com/MottainaiCI/mottainai-server/pkg/queues"
	msetting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	"github.com/mudler/anagent"
)

type DefaultTaskScheduler struct {
	Config    *setting.Config
	Scheduler *anagent.Anagent
	Fetcher   client.HttpClient

	Mutex sync.Mutex

	DefaultQueue string
	Agents       map[string]nodes.Node
}

func NewDefaultTaskScheduler(config *setting.Config, agent *anagent.Anagent) *DefaultTaskScheduler {
	ans := &DefaultTaskScheduler{
		Config:       config,
		Scheduler:    agent,
		DefaultQueue: "general",
		Agents:       make(map[string]nodes.Node, 0),
		Mutex:        sync.Mutex{},
	}

	// Initialize fetcher
	fetcher := client.NewTokenClient(
		config.GetWeb().AppURL,
		config.GetScheduler().ApiKey,
		config.ToMottainaiConfig(),
	)

	ans.Fetcher = fetcher

	return ans
}

func (s *DefaultTaskScheduler) Setup() error {
	// Initialize list of the nodes
	err := s.RetrieveNodes()
	if err != nil {
		return err
	}

	err = s.RetrieveDefaultQueue()
	if err != nil {
		return err
	}

	return nil
}

func (s *DefaultTaskScheduler) RetrieveDefaultQueue() error {
	var tlist []msetting.Setting

	req := &schema.Request{
		Route:  v1.Schema.GetSettingRoute("show_all"),
		Target: &tlist,
	}

	err := s.Fetcher.Handle(req)
	if err != nil {
		return err
	}

	// Retrieve general queue config
	for _, i := range tlist {
		if i.Key == msetting.SYSTEM_TASKS_DEFAULT_QUEUE {
			s.DefaultQueue = i.Value
			break
		}
	}

	return nil
}

func (s *DefaultTaskScheduler) RetrieveNodes() error {
	var (
		n        []nodes.Node
		filtered []nodes.Node
	)

	req := &schema.Request{
		Route:  v1.Schema.GetNodeRoute("show_all"),
		Target: &n,
	}

	err := s.Fetcher.Handle(req)
	if err != nil {
		return err
	}

	if len(s.Config.GetScheduler().Queues) > 0 {
		for _, node := range n {
			valid := false
			for _, q := range s.Config.GetScheduler().Queues {
				if node.HasQueue(q) {
					valid = true
					break
				}
			}
			if valid {
				filtered = append(filtered, node)
			}
		}
		n = filtered
	}

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	m := make(map[string]nodes.Node, 0)

	for _, node := range n {
		m[fmt.Sprintf("%s-%s", node.NodeID, node.Key)] = node
	}
	// Store agent data
	s.Agents = m

	return nil
}

func (s *DefaultTaskScheduler) GetQueues() ([]queues.Queue, error) {
	var n []queues.Queue

	req := &schema.Request{
		Route:  v1.Schema.GetQueueRoute("show_all"),
		Target: &n,
	}

	if len(s.Config.GetScheduler().Queues) > 0 {
		// POST: Get only defined queues
		body := map[string]interface{}{
			"queues": s.Config.GetScheduler().Queues,
		}

		b, err := json.Marshal(body)
		if err != nil {
			return n, err
		}

		req.Body = bytes.NewBuffer(b)
	}

	err := s.Fetcher.Handle(req)
	if err != nil {
		return n, err
	}

	if len(s.Config.GetScheduler().ExcludedQueues) > 0 && len(n) > 0 {
		filtered := []queues.Queue{}
		for _, q := range n {
			if utils.ArrayContainsString(s.Config.GetScheduler().ExcludedQueues, q.Name) {
				continue
			}
			filtered = append(filtered, q)
		}
		n = filtered
	}

	return n, nil
}

func (s *DefaultTaskScheduler) GetNodeQueues() ([]queues.NodeQueues, error) {
	var (
		n        []queues.NodeQueues
		filtered []queues.NodeQueues
	)

	// TODO: Add filter options
	req := &schema.Request{
		Route:  v1.Schema.GetNodeQueueRoute("show_all"),
		Target: &n,
	}

	err := s.Fetcher.Handle(req)
	if err != nil {
		return n, err
	}

	// Get The queues of the nodes available
	for _, q := range n {
		if _, ok := s.Agents[fmt.Sprintf("%s-%s", q.NodeId, q.AgentKey)]; ok {
			filtered = append(filtered, q)
		}
	}

	return filtered, nil
}

func (s *DefaultTaskScheduler) Schedule() error {
	tasksmap, err := s.GetTasks2Inject()
	if err != nil {
		return err
	}

	if len(tasksmap) > 0 {
		for nodeid, m := range tasksmap {

			akey := s.Agents[nodeid].Key
			nid := s.Agents[nodeid].NodeID

			for q, tasks := range m {
				fields := strings.Split(q, "|")
				// fields[0] contains queue name
				// fields[1] contains queue id
				for _, tid := range tasks {
					_, err := s.Fetcher.NodeQueueAddTask(akey, nid, fields[0], tid)
					if err != nil {
						return err
					}

					// Remote task from queue to avoid reinjection
					req := &schema.Request{
						Route: v1.Schema.GetQueueRoute("del_task"),
						Options: map[string]interface{}{
							":qid": fields[1],
							":tid": tid,
						},
					}
					resp, err := s.Fetcher.HandleAPIResponse(req)
					if err != nil {
						return err
					}
					if resp.Status == "ko" {
						return errors.New("Error on delete task " + tid +
							" from queue " + fields[0])
					}
				}
			}
		}
	}

	return nil
}

func (s *DefaultTaskScheduler) GetTasks2Inject() (map[string]map[string][]string, error) {
	ans := make(map[string]map[string][]string, 0)

	allQueues, err := s.GetQueues()
	if err != nil {
		return ans, err
	}

	queuesWithTasks := []queues.Queue{}
	// Identify queues with tasks in waiting
	for _, q := range allQueues {
		if len(q.Waiting) > 0 || len(q.PipelinesInProgress) > 0 {
			queuesWithTasks = append(queuesWithTasks, q)
		}
	}

	if len(queuesWithTasks) == 0 {
		// Nothing to do
		fmt.Println("No tasks or pipeline available. Nothing to do.")
		return ans, nil
	}

	// Retrieve node quests
	nodeQueues, err := s.GetNodeQueues()
	if err != nil {
		return ans, err
	}

	for _, q := range queuesWithTasks {
		if len(q.Waiting) > 0 {
			m, err := s.elaborateQueue(q, nodeQueues)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error on elaborate queue %s: %s",
					q, err.Error()))
				continue
			}

			if len(m) > 0 {
				for idnode, tasks := range m {
					if _, ok := ans[idnode]; !ok {
						ans[idnode] = make(map[string][]string, 0)
					}
					ans[idnode][fmt.Sprintf("%s|%s", q.Name, q.Qid)] = tasks
				}
			} else {
				fmt.Println(fmt.Sprintf(
					"No agents available for queue %s. I will try later.",
					q.Name))
			}
		}

		if len(q.PipelinesInProgress) > 0 {
			for _, pid := range q.PipelinesInProgress {
				err = s.AnalyzePipeline(pid, q)
				if err != nil {
					fmt.Println("Error on analyze pipeline " + pid + ": " + err.Error())
				}
			}
		}

	}

	return ans, nil
}

func (s *DefaultTaskScheduler) FailTask(tid, errmsg string) error {

	req := &schema.Request{
		Route: v1.Schema.GetTaskRoute("update"),
		Options: map[string]interface{}{
			"id":     tid,
			"result": msetting.TASK_RESULT_FAILED,
			"status": msetting.TASK_STATE_DONE,
		},
	}

	_, err := s.Fetcher.HandleAPIResponse(req)
	if err != nil {
		return err
	}

	req = &schema.Request{
		Route: v1.Schema.GetTaskRoute("append"),
		Options: map[string]interface{}{
			"id":     tid,
			"output": errmsg,
		},
	}

	_, err = s.Fetcher.HandleAPIResponse(req)
	if err != nil {
		return err
	}

	return nil
}

func (s *DefaultTaskScheduler) addTask2Queue(qid, tid string) error {
	req := &schema.Request{
		Route: v1.Schema.GetQueueRoute("add_task"),
		Options: map[string]interface{}{
			":qid": qid,
			":tid": tid,
		},
	}

	_, err := s.Fetcher.HandleAPIResponse(req)

	return err
}

func (s *DefaultTaskScheduler) AnalyzePipeline(pid string, q queues.Queue) error {
	allTasksDone := true
	var p tasks.Pipeline

	// Retrieve pipeline data
	req := &schema.Request{
		Route: v1.Schema.GetTaskRoute("pipeline_show"),
		Options: map[string]interface{}{
			":id": pid,
		},
		Target: &p,
	}

	err := s.Fetcher.Handle(req)
	if err != nil {
		return err
	}

	if len(p.Chain) > 0 {
		// POST: Chain pipeline
		// I need check if there is a new task to inject
		pipelineInError := false

		for _, t := range p.Chain {

			if pipelineInError {
				// POST: set the task in error
				err = s.FailTask(p.Tasks[t].ID,
					"Task in error a cause to errors with other "+
						"tasks of the pipeline: "+err.Error())
			}

			if p.Tasks[t].Status == msetting.TASK_STATE_RUNNING {
				// POST: nothing to do. we wait for the end.
				allTasksDone = false
				break
			}

			if p.Tasks[t].Status == msetting.TASK_STATE_STOPPED ||
				p.Tasks[t].Status == msetting.TASK_STATE_ASK_STOP ||
				(p.Tasks[t].Status == msetting.TASK_STATE_DONE &&
					(p.Tasks[t].Result == msetting.TASK_RESULT_ERROR ||
						p.Tasks[t].Result == msetting.TASK_RESULT_FAILED)) {
				pipelineInError = true
				allTasksDone = false
				continue
			}

			if p.Tasks[t].Status == msetting.TASK_STATE_WAIT {
				allTasksDone = false
				if !q.HasTaskInWaiting(p.Tasks[t].ID) &&
					!q.HasTaskInWaiting(p.Tasks[t].ID) {
					// POST: The task
					err = s.addTask2Queue(q.Qid, p.Tasks[t].ID)
					if err != nil {
						fmt.Println("Error on add task " + p.Tasks[t].ID +
							" in queue " + q.Qid)
					}
				}
				break
			}

		}

	} else if len(p.Chord) > 0 {
		// POST: Chord pipeline
		//       I need wait for all tasks in group before run
		//       the finals tasks.

		pipelineInError := false

		for _, t := range p.Group {
			if p.Tasks[t].Status == msetting.TASK_STATE_WAIT {
				allTasksDone = false
				break
			} else if p.Tasks[t].Result == msetting.TASK_RESULT_FAILED ||
				p.Tasks[t].Result == msetting.TASK_RESULT_ERROR {
				pipelineInError = true
				break
			}
		}

		if pipelineInError {
			// POST: Set in error all chord tasks
			for _, t := range p.Chord {
				err = s.FailTask(p.Tasks[t].ID,
					"Task in error a cause to errors with other "+
						"tasks of the pipeline: "+err.Error())
			}
		} else if allTasksDone {
			// POST: we need run the chord tasks

			for _, t := range p.Chord {

				if pipelineInError {
					// POST: set the task in error
					err = s.FailTask(p.Tasks[t].ID,
						"Task in error a cause to errors with other "+
							"tasks of the pipeline: "+err.Error())
				}

				if p.Tasks[t].Status == msetting.TASK_STATE_RUNNING {
					// POST: nothing to do. we wait for the end.
					allTasksDone = false
					break
				}

				if p.Tasks[t].Status == msetting.TASK_STATE_STOPPED ||
					p.Tasks[t].Status == msetting.TASK_STATE_ASK_STOP ||
					(p.Tasks[t].Status == msetting.TASK_STATE_DONE &&
						(p.Tasks[t].Result == msetting.TASK_RESULT_ERROR ||
							p.Tasks[t].Result == msetting.TASK_RESULT_FAILED)) {
					pipelineInError = true
					allTasksDone = false
					continue
				}

				if p.Tasks[t].Status == msetting.TASK_STATE_WAIT {
					allTasksDone = false
					if !q.HasTaskInWaiting(p.Tasks[t].ID) &&
						!q.HasTaskInWaiting(p.Tasks[t].ID) {
						// POST: The task
						err = s.addTask2Queue(q.Qid, p.Tasks[t].ID)
						if err != nil {
							fmt.Println("Error on add task " + p.Tasks[t].ID +
								" in queue " + q.Qid)
						}
					}
					break
				}
			}
		}

	} else {
		// POST: Groups pipeline
		//       If all tasks are completed i can delete the pipeline from the queue

		for _, t := range p.Group {
			if p.Tasks[t].Status == msetting.TASK_STATE_WAIT {
				allTasksDone = false
				break
			}
		}

	}

	if allTasksDone {
		req := &schema.Request{
			Route: v1.Schema.GetQueueRoute("del_pipeline_in_progress"),
			Options: map[string]interface{}{
				":qid": q.Qid,
				":pid": pid,
			},
		}

		_, err := s.Fetcher.HandleAPIResponse(req)
		if err != nil {
			if req.Response != nil {
				fmt.Println("ERROR: ", req.Response.StatusCode)
				fmt.Println(string(req.ResponseRaw))
			}
			return err
		}
	}

	return nil
}

func (s *DefaultTaskScheduler) elaborateQueue(queue queues.Queue, nodeQueues []queues.NodeQueues) (map[string][]string, error) {
	ans := make(map[string][]string, 0)

	isDefaultQueue := false
	if queue.Name == s.DefaultQueue {
		isDefaultQueue = true
	}

	// Retrieve the list of agents with the specified queues
	validAgents := []specs.NodeSlots{}
	for _, node := range nodeQueues {
		nodeKey := fmt.Sprintf("%s-%s", node.NodeId, node.AgentKey)
		if maxTasks, ok := s.Agents[nodeKey].Queues[queue.Name]; ok {
			tt, ok := node.Queues[queue.Name]
			slot := specs.NodeSlots{
				Key: nodeKey,
			}

			if !ok {
				slot.AvailableSlot = maxTasks
				validAgents = append(validAgents, slot)
			} else if len(tt) < maxTasks {
				slot.AvailableSlot = maxTasks - len(tt)
				validAgents = append(validAgents, slot)
			}
		} else if isDefaultQueue {
			slot := specs.NodeSlots{
				Key: nodeKey,
			}
			// POST:Push the task only
			if !s.Agents[nodeKey].Standalone {
				// Check if there are already task in queue
				if qtasks, ok := node.Queues[queue.Name]; ok {
					if len(qtasks) < s.Agents[nodeKey].Concurrency {
						slot.AvailableSlot = s.Agents[nodeKey].Concurrency - len(qtasks)
						validAgents = append(validAgents, slot)
					}
				} else {
					// POST: the node doesn't contains queue tasks for the
					//       default queue
					slot.AvailableSlot = s.Agents[nodeKey].Concurrency
					validAgents = append(validAgents, slot)
				}
			}
		}
	}

	if len(validAgents) > 0 {
		sort.Sort(specs.NodeSlotsList(validAgents))

		for _, node := range validAgents {
			if len(queue.Waiting) == 0 {
				break
			}

			for ; node.AvailableSlot > 0; node.AvailableSlot-- {
				if len(queue.Waiting) == 0 {
					break
				}
				tid := queue.Waiting[0]
				queue.Waiting = queue.Waiting[1:]
				if _, ok := ans[node.Key]; ok {
					ans[node.Key] = append(ans[node.Key], tid)
				} else {
					ans[node.Key] = []string{tid}
				}
			}
		}
	}

	return ans, nil
}

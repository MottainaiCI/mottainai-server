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
	"errors"
	"fmt"
	"sort"
	"strings"

	setting "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/config"
	specs "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/specs"
	"github.com/MottainaiCI/mottainai-server/pkg/queues"
	msetting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	"github.com/mudler/anagent"
)

type SimpleTaskScheduler struct {
	*DefaultTaskScheduler
}

func NewSimpleTaskScheduler(config *setting.Config, agent *anagent.Anagent) *SimpleTaskScheduler {
	return &SimpleTaskScheduler{
		DefaultTaskScheduler: NewDefaultTaskScheduler(config, agent),
	}
}

func (s *SimpleTaskScheduler) Schedule() error {
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

					fmt.Println("Assigned task " + tid + " to agent " + nid + ".")

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

func (s *SimpleTaskScheduler) GetTasks2Inject() (map[string]map[string][]string, error) {
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
				fmt.Println(fmt.Sprintf(
					"Analyzing pipeline %s of the queue %s",
					pid, q.Name,
				))
				err = s.AnalyzePipeline(pid, q, allQueues)
				if err != nil {
					fmt.Println("Error on analyze pipeline " + pid + ": " + err.Error())
				}
			}
		}

	}

	return ans, nil
}

func (s *SimpleTaskScheduler) AnalyzePipeline(pid string, q queues.Queue, allQueues []queues.Queue) error {
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
		allTasksDone, err = s.elaboratePipelineChain(&p, q, allQueues)

	} else if len(p.Chord) > 0 {
		// POST: Chord pipeline
		//       I need wait for all tasks in group before run
		//       the finals tasks.
		allTasksDone, err = s.elaboratePipelineChord(&p, q, allQueues)

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

		fmt.Println(fmt.Sprintf("Pipeline %s completed.", pid))

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

		fmt.Println(fmt.Sprintf("Delete pipeline %s from queue %s (%s)",
			pid, q.Name, q.Qid))

		// Update end time of the pipeline
		req = &schema.Request{
			Route: v1.Schema.GetTaskRoute("pipeline_completed"),
			Options: map[string]interface{}{
				":id": pid,
			},
		}

		_, err = s.Fetcher.HandleAPIResponse(req)
		if err != nil {
			if req.Response != nil {
				fmt.Println("ERROR: ", req.Response.StatusCode)
				fmt.Println(string(req.ResponseRaw))
			}
			return err
		}
	} else {
		fmt.Println(fmt.Sprintf("Pipeline %s running. Nothing to do.", pid))
	}

	return nil
}

func (s *DefaultTaskScheduler) elaboratePipelineChord(p *tasks.Pipeline, q queues.Queue, allQueues []queues.Queue) (bool, error) {
	var err error
	var errGroup error
	var errChord error

	pipelineInError := false
	allTasksDone := true

	for _, t := range p.Group {
		if p.Tasks[t].Status == msetting.TASK_STATE_WAIT || p.Tasks[t].Status == msetting.TASK_STATE_RUNNING {
			allTasksDone = false
			break
		} else if p.Tasks[t].Result == msetting.TASK_RESULT_FAILED ||
			p.Tasks[t].Result == msetting.TASK_RESULT_ERROR {
			errGroup = errors.New(fmt.Sprintf("Task %s in failed.", p.Tasks[t].ID))
			pipelineInError = true
			break
		}
	}

	if pipelineInError {
		// POST: Set in error all chord tasks
		for _, t := range p.Chord {
			if p.Tasks[t].Status == msetting.TASK_STATE_WAIT {
				if errGroup != nil {
					err = s.FailTask(p.Tasks[t].ID, p.Tasks[t].Queue,
						"Task in error a cause to errors with other "+
							"tasks of the pipeline: "+errGroup.Error())
				} else {
					err = s.FailTask(p.Tasks[t].ID, p.Tasks[t].Queue,
						"Task in error a cause to errors with other "+
							"tasks of the pipeline.")
				}
			}
		}
	} else if allTasksDone {
		// POST: we need run the chord tasks

		for _, t := range p.Chord {

			if pipelineInError {

				if p.Tasks[t].Result != msetting.TASK_RESULT_FAILED &&
					p.Tasks[t].Result != msetting.TASK_RESULT_ERROR {
					errMsg := "."
					if errGroup != nil {
						errMsg = fmt.Sprintf(": %s", errGroup.Error())
					} else if errChord != nil {
						errMsg = fmt.Sprintf(": %s", errChord.Error())
					}

					// POST: set the task in error
					err = s.FailTask(p.Tasks[t].ID, p.Tasks[t].Queue,
						"Task in error a cause to errors with other "+
							"tasks of the pipeline"+errMsg)
				}
				// else I already set the task in error.

				continue
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
				errChord = errors.New(fmt.Sprintf(
					"Task %s failed.", p.Tasks[t].ID,
				))
				continue
			}

			if p.Tasks[t].Status == msetting.TASK_STATE_WAIT {
				allTasksDone = false
				if p.Tasks[t].Queue == q.Name &&
					!q.HasTaskInWaiting(p.Tasks[t].ID) &&
					!q.HasTaskInRunning(p.Tasks[t].ID) {
					// POST: The task
					err = s.AddTask2Queue(p.Tasks[t].ID, p.Tasks[t].Queue)
					if err != nil {
						fmt.Println("Error on add task " + p.Tasks[t].ID +
							" in queue " + p.Tasks[t].Queue)
					}

					fmt.Println(fmt.Sprintf(
						"For pipeline %s added task %s to queue (%s).",
						p.ID, p.Tasks[t].ID, p.Tasks[t].Queue))

				} else if p.Tasks[t].Queue != q.Name {

					// Retrieve queue data
					var qTask *queues.Queue = nil

					for idx, qT := range allQueues {
						if qT.Name == p.Tasks[t].Queue {
							qTask = &allQueues[idx]
							break
						}
					}

					if qTask == nil || (!qTask.HasTaskInWaiting(p.Tasks[t].ID) && !qTask.HasTaskInRunning(p.Tasks[t].ID)) {
						// POST: If qTask is nil means that the queue is not present yet.
						err = s.AddTask2Queue(p.Tasks[t].ID, p.Tasks[t].Queue)
						if err != nil {
							fmt.Println("Error on add task " + p.Tasks[t].ID +
								" in queue " + p.Tasks[t].Queue)
						} else {
							fmt.Println(fmt.Sprintf("For pipeline %s added task %s to queue (%s).",
								p.ID, p.Tasks[t].ID, p.Tasks[t].Queue))
						}
					}
				}
				break
			}
		}
	}

	return allTasksDone, err
}

func (s *DefaultTaskScheduler) elaboratePipelineChain(p *tasks.Pipeline, q queues.Queue, allQueues []queues.Queue) (bool, error) {
	var err error

	allTasksDone := true

	// POST: Chain pipeline
	// I need check if there is a new task to inject
	pipelineInError := false

	for _, t := range p.Chain {

		fmt.Println(fmt.Sprintf(
			"For pipeline %s the task %s is in state %s with result %s.",
			p.ID, p.Tasks[t].ID, p.Tasks[t].Status, p.Tasks[t].Result,
		))

		if pipelineInError {
			// POST: set the task in error
			err = s.FailTask(p.Tasks[t].ID, p.Tasks[t].Queue,
				"Task in error a cause to errors with other "+
					"tasks of the pipeline")

			continue
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
			continue
		}

		if p.Tasks[t].Status == msetting.TASK_STATE_WAIT {
			allTasksDone = false

			if p.Tasks[t].Queue == q.Name &&
				!q.HasTaskInWaiting(p.Tasks[t].ID) &&
				!q.HasTaskInRunning(p.Tasks[t].ID) {
				// POST: The task
				err = s.AddTask2Queue(p.Tasks[t].ID, p.Tasks[t].Queue)
				if err != nil {
					fmt.Println("Error on add task " + p.Tasks[t].ID +
						" in queue " + p.Tasks[t].Queue)
				} else {
					fmt.Println(fmt.Sprintf("For pipeline %s added task %s to queue (%s).",
						p.ID, p.Tasks[t].ID, p.Tasks[t].Queue))
				}
			} else if p.Tasks[t].Queue != q.Name {
				// Retrieve queue data
				var qTask *queues.Queue = nil

				for idx, qT := range allQueues {
					if qT.Name == p.Tasks[t].Queue {
						qTask = &allQueues[idx]
						break
					}
				}

				if qTask == nil || (!qTask.HasTaskInWaiting(p.Tasks[t].ID) && !qTask.HasTaskInRunning(p.Tasks[t].ID)) {
					// POST: If qTask is nil means that the queue is not present yet.
					err = s.AddTask2Queue(p.Tasks[t].ID, p.Tasks[t].Queue)
					if err != nil {
						fmt.Println("Error on add task " + p.Tasks[t].ID +
							" in queue " + p.Tasks[t].Queue)
					} else {
						fmt.Println(fmt.Sprintf("For pipeline %s added task %s to queue (%s).",
							p.ID, p.Tasks[t].ID, p.Tasks[t].Queue))
					}
				}

			}
			break
		}

	}

	return allTasksDone, err
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

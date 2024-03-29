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
	"fmt"
	"sync"

	setting "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/config"
	"github.com/MottainaiCI/mottainai-server/pkg/client"
	"github.com/MottainaiCI/mottainai-server/pkg/nodes"
	"github.com/MottainaiCI/mottainai-server/pkg/queues"
	msetting "github.com/MottainaiCI/mottainai-server/pkg/settings"
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

func (s *DefaultTaskScheduler) FailTask(tid, queue, errmsg string) error {

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

	// Delete task from the queue
	err = s.DelTask2Queue(tid, queue)
	if err != nil {
		return err
	}

	return nil
}

func (s *DefaultTaskScheduler) DelTask2Queue(tid, queue string) error {
	var qid string

	if queue == "" {
		queue = s.DefaultQueue
	}

	// Retrieve qid of the queue
	req := &schema.Request{
		Route: v1.Schema.GetQueueRoute("get_qid"),
		Options: map[string]interface{}{
			":name": queue,
		},
		Target: &qid,
	}

	err := s.Fetcher.Handle(req)
	if err != nil {
		return err
	}

	req = &schema.Request{
		Route: v1.Schema.GetQueueRoute("del_task"),
		Options: map[string]interface{}{
			":qid": qid,
			":tid": tid,
		},
	}

	_, err = s.Fetcher.HandleAPIResponse(req)

	fmt.Println(fmt.Sprintf("Deleted task %s from queue %s", tid, qid))

	return err

}

func (s *DefaultTaskScheduler) AddTask2Queue(tid, queue string) error {
	var qid string

	if queue == "" {
		queue = s.DefaultQueue
	}

	// Retrieve qid of the queue
	req := &schema.Request{
		Route: v1.Schema.GetQueueRoute("get_qid"),
		Options: map[string]interface{}{
			":name": queue,
		},
		Target: &qid,
	}

	err := s.Fetcher.Handle(req)
	if err != nil {
		return err
	}

	req = &schema.Request{
		Route: v1.Schema.GetQueueRoute("add_task"),
		Options: map[string]interface{}{
			":qid": qid,
			":tid": tid,
		},
	}

	_, err = s.Fetcher.HandleAPIResponse(req)

	return err
}

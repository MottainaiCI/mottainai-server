/*

Copyright (C) 2021  Daniele Rondina, geaaru@sabayonlinux.org

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
	"bytes"
	"encoding/json"

	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	queues "github.com/MottainaiCI/mottainai-server/pkg/queues"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
)

func (d *Fetcher) NodeQueueCreate(agentKey, nodeId string, queues map[string][]string) (event.APIResponse, error) {

	req := &schema.Request{
		Route: v1.Schema.GetNodeQueueRoute("create"),
	}

	msg := map[string]interface{}{
		"akey":   agentKey,
		"nodeid": nodeId,
		"queues": queues,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return event.APIResponse{}, err
	}

	req.Body = bytes.NewBuffer(b)
	return d.HandleAPIResponse(req)
}

func (d *Fetcher) NodeQueueDelete(agentKey, nodeId string) (event.APIResponse, error) {

	req := &schema.Request{
		Route: v1.Schema.GetNodeQueueRoute("delete"),
	}

	msg := map[string]interface{}{
		"akey":   agentKey,
		"nodeid": nodeId,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return event.APIResponse{}, err
	}

	req.Body = bytes.NewBuffer(b)
	return d.HandleAPIResponse(req)
}

func (d *Fetcher) NodeQueueAddTask(agentKey, nodeId, queue, taskid string) (event.APIResponse, error) {

	req := &schema.Request{
		Route: v1.Schema.GetNodeQueueRoute("add_task"),
		Options: map[string]interface{}{
			":queue": queue,
			":tid":   taskid,
		},
	}

	msg := map[string]interface{}{
		"akey":   agentKey,
		"nodeid": nodeId,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return event.APIResponse{}, err
	}

	req.Body = bytes.NewBuffer(b)
	return d.HandleAPIResponse(req)
}

func (f *Fetcher) NodeQueueDelById(nodeQueueId string) (event.APIResponse, error) {
	req := &schema.Request{
		Route: v1.Schema.GetNodeQueueRoute("delete_byid"),
		Options: map[string]interface{}{
			":id": nodeQueueId,
		},
	}

	return f.HandleAPIResponse(req)
}

func (d *Fetcher) NodeQueueDelTask(agentKey, nodeId, queue, taskid string) (event.APIResponse, error) {

	req := &schema.Request{
		Route: v1.Schema.GetNodeQueueRoute("del_task"),
		Options: map[string]interface{}{
			":queue": queue,
			":tid":   taskid,
		},
	}

	msg := map[string]interface{}{
		"akey":   agentKey,
		"nodeid": nodeId,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return event.APIResponse{}, err
	}

	req.Body = bytes.NewBuffer(b)

	return d.HandleAPIResponse(req)
}

func (d *Fetcher) QueueCreate(name string) (event.APIResponse, error) {
	req := &schema.Request{
		Route: v1.Schema.GetQueueRoute("create"),
		Options: map[string]interface{}{
			":name": name,
		},
	}

	return d.HandleAPIResponse(req)
}

func (d *Fetcher) QueueDelete(qid string) (event.APIResponse, error) {
	req := &schema.Request{
		Route: v1.Schema.GetQueueRoute("delete"),
		Options: map[string]interface{}{
			":qid": qid,
		},
	}

	return d.HandleAPIResponse(req)
}

func (d *Fetcher) QueueGetQid(name string) (string, error) {
	var qid string

	req := &schema.Request{
		Route: v1.Schema.GetQueueRoute("get_qid"),
		Options: map[string]interface{}{
			":name": name,
		},
		Target: &qid,
	}

	err := d.Handle(req)
	return qid, err
}

func (d *Fetcher) QueueAddTaskInProgress(qid, taskid string) (event.APIResponse, error) {
	req := &schema.Request{
		Route: v1.Schema.GetQueueRoute("add_task_in_progress"),
		Options: map[string]interface{}{
			":qid": qid,
			":tid": taskid,
		},
	}

	return d.HandleAPIResponse(req)
}

func (d *Fetcher) QueueDelTaskInProgress(qid, taskid string) (event.APIResponse, error) {
	req := &schema.Request{
		Route: v1.Schema.GetQueueRoute("del_task_in_progress"),
		Options: map[string]interface{}{
			":qid": qid,
			":tid": taskid,
		},
	}

	return d.HandleAPIResponse(req)
}

func (d *Fetcher) NodeQueueGetTasks(id string) (queues.NodeQueues, error) {
	var n queues.NodeQueues
	req := &schema.Request{
		Route: v1.Schema.GetNodeQueueRoute("show"),
		Options: map[string]interface{}{
			":id": id,
		},
		Target: &n,
	}
	err := d.Handle(req)

	return n, err
}

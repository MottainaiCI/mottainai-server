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

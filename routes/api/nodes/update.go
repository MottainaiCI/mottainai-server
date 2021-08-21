/*

Copyright (C) 2017-2021  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>
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

package nodesapi

import (
	"errors"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	nodes "github.com/MottainaiCI/mottainai-server/pkg/nodes"
)

type NodeUpdate struct {
	NodeID      string         `json:"nodeid" form:"nodeid"`
	Key         string         `json:"key" form:"key"`
	Hostname    string         `json:"hostname" form:"hostname"`
	Standalone  bool           `json:"standalone" form:"standalone"`
	Queues      map[string]int `json:"queues" form:"queues"`
	Concurrency int            `json:"concurrency" form:"concurrency"`
}

func Register(nodedata NodeUpdate, ctx *context.Context, db *database.Database) error {
	key := nodedata.Key
	nodeid := nodedata.NodeID
	hostname := nodedata.Hostname

	if len(key) == 0 {
		return errors.New("Invalid key")
	}

	n := db.Driver.AllNodes()

	nodefound, err := db.Driver.GetNodeByKey(key)
	if err != nil {
		ctx.APIActionFailed("", "", "Node not found", "", 404)
		return nil
	}

	hb := time.Now().UTC().Format("20060102150405")
	doc := map[string]interface{}{
		"nodeid":      nodeid,
		"hostname":    hostname,
		"last_report": hb,
		"standalone":  nodedata.Standalone,
	}

	if !nodefound.OverrideQueues {
		doc["queues"] = nodedata.Queues
		doc["concurrency"] = nodedata.Concurrency
	}

	// Find my position between nodes

	activeNodes := []nodes.Node{}

	for _, i := range n {
		if i.LastReport == "" {
			continue
		}
		activeNodes = append(activeNodes, i)
	}

	var pos int
	for p, i := range activeNodes {
		if i.ID == nodefound.ID {
			pos = p
		}
	}

	db.Driver.UpdateNode(nodefound.ID, doc)

	// Chech if there are tasks in queue
	task_in_queue := false
	nq, err := db.Driver.GetNodeQueuesByKey(key, nodeid)
	if err == nil && nq.ID == "" {
		// POST: check if we need to create the node queue
		q := make(map[string][]string, 0)

		for k, _ := range doc["queues"].(map[string]int) {
			q[k] = []string{}
		}

		_, err = db.Driver.CreateNodeQueues(map[string]interface{}{
			"akey":          key,
			"nodeid":        nodeid,
			"queues":        q,
			"creation_date": hb,
		})

		if err != nil {
			return errors.New("Error on create node queue: " + err.Error())
		}

	} else if err != nil {
		return err
	} else {

		if len(nq.Queues) > 0 {
			for _, tt := range nq.Queues {
				if len(tt) > 0 {
					task_in_queue = true
					break
				}
			}
		}
	}

	resp := &nodes.NodeRegisterResponse{
		NumNodes:     len(activeNodes),
		Position:     pos,
		TaskInQueue:  task_in_queue,
		NodeUniqueId: nq.ID,
	}

	ctx.APIEventReport(event.APIResponse{
		Data:      resp.ToJson(),
		Processed: "true",
		Status:    "ok",
	}, 200)
	return nil
}

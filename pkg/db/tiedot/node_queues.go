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
package tiedot

import (
	"errors"
	"strconv"
	"sync"

	"github.com/MottainaiCI/mottainai-server/pkg/queues"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
)

var NodeQueuesColl = "NodeQueues"
var NodeQueuesMutex = sync.Mutex{}

func (d *Database) IndexNodeQueue() {
	d.AddIndex(NodeQueuesColl, []string{"akey"})
	d.AddIndex(NodeQueuesColl, []string{"nodeid"})
}

func (d *Database) CreateNodeQueues(t map[string]interface{}) (string, error) {
	return d.InsertDoc(NodeQueuesColl, t)
}

func (d *Database) InsertNodeQueues(q *queues.NodeQueues) (string, error) {
	return d.CreateNodeQueues(q.ToMap())
}

func (d *Database) DeleteNodeQueues(docId string) error {
	return d.DeleteDoc(NodeQueuesColl, docId)
}

func (d *Database) UpdateNodeQueues(docId string, t map[string]interface{}) error {
	return d.UpdateDoc(NodeQueuesColl, docId, t)
}

func (d *Database) GetNodeQueues(docId string) (queues.NodeQueues, error) {
	doc, err := d.GetDoc(NodeQueuesColl, docId)
	if err != nil {
		return queues.NodeQueues{}, err
	}

	t := queues.NewNodeQueuesFromMap(doc)
	t.ID = docId
	return t, err
}

func (d *Database) AddNodeQueuesTask(agentKey, nodeid, queue, taskid string) error {
	NodeQueuesMutex.Lock()
	defer NodeQueuesMutex.Unlock()

	nq, err := d.GetNodeQueuesByKey(agentKey, nodeid)
	if err != nil {
		return err
	}

	if _, ok := nq.Queues[queue]; !ok {
		nq.Queues[queue] = []string{}
	}

	nq.Queues[queue] = append(nq.Queues[queue], taskid)

	err = d.UpdateNodeQueues(nq.ID, map[string]interface{}{
		"akey":          nq.AgentKey,
		"nodeid":        nq.NodeId,
		"queues":        nq.Queues,
		"creation_date": nq.CreationDate,
	})

	return err
}

func (d *Database) DelNodeQueuesTask(agentKey, nodeid, queue, taskid string) error {
	NodeQueuesMutex.Lock()
	defer NodeQueuesMutex.Unlock()

	nq, err := d.GetNodeQueuesByKey(agentKey, nodeid)
	if err != nil {
		return err
	}

	if nq.ID == "" {
		return errors.New("Node queue not found")
	}

	if _, ok := nq.Queues[queue]; ok {
		tasks := nq.Queues[queue]

		ntasks := []string{}

		for _, t := range tasks {
			if t == taskid {
				continue
			}
			ntasks = append(ntasks, t)
		}

		nq.Queues[queue] = ntasks
	}

	err = d.UpdateNodeQueues(nq.ID, map[string]interface{}{
		"akey":          nq.AgentKey,
		"nodeid":        nq.NodeId,
		"queues":        nq.Queues,
		"creation_date": nq.CreationDate,
	})

	return err
}

func (d *Database) GetNodeQueuesByKey(agentKey, nodeid string) (queues.NodeQueues, error) {
	var res []queues.NodeQueues

	queuesFound, err := d.FindDoc(NodeQueuesColl,
		`{ "n":[{"eq": "`+nodeid+`", "in": ["nodeid"]}, {"eq": "`+agentKey+`", "in": ["akey"]}]}`)

	if err != nil || len(queuesFound) != 1 {
		return queues.NodeQueues{}, err
	}

	for docid := range queuesFound {
		q, err := d.GetNodeQueues(docid)
		q.ID = docid
		if err != nil {
			return queues.NodeQueues{}, err
		}
		res = append(res, q)
	}

	return res[0], nil
}

func (d *Database) ListNodeQueues() []dbcommon.DocItem {
	return d.ListDocs(NodeQueuesColl)
}

func (d *Database) AllNodesQueues() []queues.NodeQueues {
	nodec := d.DB().Use(NodeQueuesColl)
	node_list := make([]queues.NodeQueues, 0)

	nodec.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := queues.NewNodeQueueFromJson(docContent)
		t.ID = strconv.Itoa(id)
		node_list = append(node_list, t)
		return true
	})
	return node_list
}

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

package arangodb

import (
	"errors"
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

	queryResult, err := d.FindDoc("",
		`FOR c IN `+NodeQueuesColl+`
		FILTER c.nodeid == "`+nodeid+`" AND c.akey == "`+agentKey+`"
		RETURN c`)

	if err != nil || len(queryResult) != 1 {
		return queues.NodeQueues{}, err
	}

	for id, doc := range queryResult {
		n := queues.NewNodeQueuesFromMap(doc.(map[string]interface{}))
		n.ID = id
		res = append(res, n)
	}
	return res[0], nil
}

func (d *Database) ListNodeQueues() []dbcommon.DocItem {
	return d.ListDocs(NodeQueuesColl)
}

func (d *Database) AllNodesQueues() []queues.NodeQueues {
	queue_list := make([]queues.NodeQueues, 0)

	docs, err := d.FindDoc("", "FOR c IN "+NodeQueuesColl+" return c")
	if err != nil {
		return []queues.NodeQueues{}
	}

	for id, doc := range docs {
		t := queues.NewNodeQueuesFromMap(doc.(map[string]interface{}))
		t.ID = id
		queue_list = append(queue_list, t)
	}

	return queue_list
}

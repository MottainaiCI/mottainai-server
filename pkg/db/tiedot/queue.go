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

package tiedot

import (
	"errors"
	"strconv"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/queues"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
)

var QueueColl = "Queues"

func (d *Database) IndexQueue() {
	d.AddIndex(QueueColl, []string{"qid"})
	d.AddIndex(QueueColl, []string{"name"})
}

func (d *Database) CreateQueue(t map[string]interface{}) (string, error) {
	return d.InsertDoc(QueueColl, t)
}

func (d *Database) InsertQueue(q *queues.Queue) (string, error) {
	return d.CreateQueue(q.ToMap())
}

func (d *Database) DeleteQueue(docId string) error {
	return d.DeleteDoc(QueueColl, docId)
}

func (d *Database) UpdateQueue(docId string, t map[string]interface{}) error {
	return d.UpdateDoc(QueueColl, docId, t)
}

func (d *Database) GetQueue(docId string) (queues.Queue, error) {
	doc, err := d.GetDoc(QueueColl, docId)
	if err != nil {
		return queues.Queue{}, err
	}

	t := queues.NewQueueFromMap(doc)
	t.ID = docId
	return t, err
}

func (d *Database) GetQueueByQid(qid string) (queues.Queue, error) {
	var res []queues.Queue

	queuesFound, err := d.FindDoc(QueueColl, `[{"eq": "`+qid+`", "in": ["qid"]}]`)
	if err != nil || len(queuesFound) != 1 {
		return queues.Queue{}, err
	}

	for docid := range queuesFound {
		q, err := d.GetQueue(docid)
		q.ID = docid
		if err != nil {
			return queues.Queue{}, err
		}
		res = append(res, q)
	}

	return res[0], nil
}

func (d *Database) GetQueueByKey(name string) (queues.Queue, error) {
	var res []queues.Queue

	queuesFound, err := d.FindDoc(QueueColl, `[{"eq": "`+name+`", "in": ["name"]}]`)
	if err != nil || len(queuesFound) != 1 {
		return queues.Queue{}, err
	}

	for docid := range queuesFound {
		q, err := d.GetQueue(docid)
		q.ID = docid
		if err != nil {
			return queues.Queue{}, err
		}
		res = append(res, q)
	}

	return res[0], nil
}

func (d *Database) ListQueues() []dbcommon.DocItem {
	return d.ListDocs(QueueColl)
}

func (d *Database) AllQueues() []queues.Queue {
	queuec := d.DB().Use(QueueColl)
	queue_list := make([]queues.Queue, 0)

	queuec.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := queues.NewFromJson(docContent)
		t.ID = strconv.Itoa(id)
		queue_list = append(queue_list, t)
		return true
	})
	return queue_list
}

func (d *Database) AddTaskInProgress2Queue(qid, taskid string) error {
	ud := time.Now().UTC().Format("20060102150405")
	// TODO: add a semaphore

	q, err := d.GetQueueByQid(qid)
	if err != nil {
		return err
	}

	tasks := q.InProgress
	ntasks := []string{}

	present := false
	for _, t := range tasks {
		if t == taskid {
			present = true
			break
		}
		ntasks = append(ntasks, t)
	}

	if present {
		return errors.New("task already present in queue")
	}

	// Check if task is in waiting. If yes i will drop it.
	tasks = q.Waiting
	wtasks := []string{}
	for _, t := range tasks {
		if t == taskid {
			continue
		}
		wtasks = append(wtasks, t)
	}

	ntasks = append(ntasks, taskid)

	m := map[string]interface{}{
		"qid":              q.Qid,
		"name":             q.Name,
		"tasks_waiting":    wtasks,
		"tasks_inprogress": ntasks,
		"creation_date":    q.CreationDate,
		"update_date":      ud,
	}

	err = d.UpdateQueue(q.ID, m)

	return err
}

func (d *Database) DelTaskInProgress2Queue(qid, taskid string) error {

	ud := time.Now().UTC().Format("20060102150405")
	// TODO: add a semaphore

	q, err := d.GetQueueByQid(qid)
	if err != nil {
		return err
	}

	tasks := q.InProgress
	ntasks := []string{}

	for _, t := range tasks {
		if t == taskid {
			continue
		}

		ntasks = append(ntasks, t)
	}

	err = d.UpdateQueue(q.ID, map[string]interface{}{
		"qid":              q.Qid,
		"name":             q.Name,
		"tasks_waiting":    q.Waiting,
		"tasks_inprogress": ntasks,
		"creation_date":    q.CreationDate,
		"update_date":      ud,
	})

	return err
}

func (d *Database) AddTaskInWaiting2Queue(qid, taskid string) error {
	ud := time.Now().UTC().Format("20060102150405")
	// TODO: add a semaphore
	// TODO: add check that the task is not already in waiting

	q, err := d.GetQueueByQid(qid)
	if err != nil {
		return err
	}

	tasks := q.Waiting
	ntasks := []string{}

	present := false
	for _, t := range tasks {
		if t == taskid {
			present = true
			break
		}
		ntasks = append(ntasks, t)
	}

	if present {
		return errors.New("task already present in queue")
	}

	ntasks = append(ntasks, taskid)

	m := map[string]interface{}{
		"qid":              q.Qid,
		"name":             q.Name,
		"tasks_waiting":    ntasks,
		"tasks_inprogress": q.InProgress,
		"creation_date":    q.CreationDate,
		"update_date":      ud,
	}

	err = d.UpdateQueue(q.ID, m)

	return err
}

func (d *Database) DelTaskInWaiting2Queue(qid, taskid string) error {
	ud := time.Now().UTC().Format("20060102150405")
	// TODO: add a semaphore

	q, err := d.GetQueueByQid(qid)
	if err != nil {
		return err
	}

	tasks := q.Waiting
	ntasks := []string{}

	for _, t := range tasks {
		if t == taskid {
			continue
		}

		ntasks = append(ntasks, t)
	}

	err = d.UpdateQueue(q.ID, map[string]interface{}{
		"qid":              q.Qid,
		"name":             q.Name,
		"tasks_waiting":    ntasks,
		"tasks_inprogress": q.InProgress,
		"creation_date":    q.CreationDate,
		"update_date":      ud,
	})

	return err
}

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

package queuesapi

import (
	"errors"
	"fmt"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
)

func Remove(q NodeQueue, ctx *context.Context, db *database.Database) error {
	doc, err := db.Driver.GetNodeQueuesByKey(q.AgentKey, q.NodeId)
	if err != nil {
		return err // Do not treat it as an error, we have no node with such id.
	}

	if doc.ID == "" {
		return errors.New("Node queue not found")
	}

	err = db.Driver.DeleteNodeQueues(doc.ID)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func RemoveById(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	if id == "" {
		return errors.New("Invalid node queue id")
	}

	doc, err := db.Driver.GetNodeQueues(id)
	if err != nil {
		return err // Do not treat it as an error, we have no node with such id.
	}

	if doc.ID == "" {
		return errors.New("Node queue not found")
	}

	err = db.Driver.DeleteNodeQueues(doc.ID)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func RemoveQueue(ctx *context.Context, db *database.Database) error {
	qid := ctx.Params(":qid")

	if qid == "" {
		return errors.New("Invalid queue id")
	}

	doc, err := db.Driver.GetQueueByQid(qid)
	if err != nil {
		return err // Do not treat it as an error, we have no node with such id.
	}

	if doc.ID == "" {
		return errors.New("Queue not found")
	}

	err = db.Driver.DeleteQueue(doc.ID)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func DelTaskInProgress(ctx *context.Context, db *database.Database) error {
	qid := ctx.Params(":qid")
	tid := ctx.Params(":tid")

	if qid == "" {
		return errors.New("Invalid queue id")
	}

	if tid == "" {
		return errors.New("Invalid task id")
	}

	err := db.Driver.DelTaskInProgress2Queue(qid, tid)

	if err != nil {
		fmt.Println(fmt.Sprintf("Error on remove task %s from queue %s: %s.",
			tid, qid, err.Error()))
		return err
	}
	fmt.Println(fmt.Sprintf("Removed task %s from queue %s.", tid, qid))

	ctx.APIActionSuccess()
	return nil
}

func DelTaskInWaiting(ctx *context.Context, db *database.Database) error {
	qid := ctx.Params(":qid")
	tid := ctx.Params(":tid")

	if qid == "" {
		return errors.New("Invalid queue id")
	}

	if tid == "" {
		return errors.New("Invalid task id")
	}

	err := db.Driver.DelTaskInWaiting2Queue(qid, tid)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func DelTask(queue NodeQueue, ctx *context.Context, db *database.Database) error {
	q := ctx.Params(":queue")
	tid := ctx.Params(":tid")

	if q == "" {
		return errors.New("Invalid queue")
	}

	if tid == "" {
		return errors.New("Invalid task id")
	}

	if queue.AgentKey == "" {
		return errors.New("Invalid agent key")
	}

	if queue.NodeId == "" {
		return errors.New("Invalid node id")
	}

	err := db.Driver.DelNodeQueuesTask(queue.AgentKey, queue.NodeId, q, tid)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error on remove task %s from node queue %s (%s, %s): %s.",
			tid, q, queue.AgentKey, queue.NodeId, err.Error()))
		return err
	}

	fmt.Println(fmt.Sprintf("Removed task %s from node queue %s (%s, %s).",
		tid, q, queue.AgentKey, queue.NodeId))
	ctx.APIActionSuccess()
	return nil
}

func DelPipelineInWaiting(ctx *context.Context, db *database.Database) error {
	qid := ctx.Params(":qid")
	pid := ctx.Params(":pid")

	if qid == "" {
		return errors.New("Invalid queue id")
	}

	if pid == "" {
		return errors.New("Invalid pipeline id")
	}

	err := db.Driver.DelPipelineInWaiting2Queue(qid, pid)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func DelPipelineInProgress(ctx *context.Context, db *database.Database) error {
	qid := ctx.Params(":qid")
	pid := ctx.Params(":pid")

	if qid == "" {
		return errors.New("Invalid queue id")
	}

	if pid == "" {
		return errors.New("Invalid pipeline id")
	}

	err := db.Driver.DelPipelineInProgress2Queue(qid, pid)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

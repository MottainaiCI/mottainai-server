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
		return err
	}

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

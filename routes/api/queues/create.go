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
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/gofrs/uuid"
)

type NodeQueue struct {
	AgentKey string              `json:"akey" form:"akey"`
	NodeId   string              `json:"nodeid" form:"nodeid"`
	Queues   map[string][]string `json:"queues,omitempty" form:"queues"`
}

func APICreate(queue NodeQueue, ctx *context.Context, db *database.Database) error {
	id, err := Create(queue, ctx, db)
	if err != nil {
		return err
	}

	ctx.APICreationSuccess(id, "queue")
	return nil
}

func APIQueueCreate(ctx *context.Context, db *database.Database) error {
	qid, err := uuid.NewV4()
	name := ctx.Params(":name")
	ct := time.Now().UTC().Format("20060102150405")

	if name == "" {
		return errors.New("Invalid queue name")
	}

	// TODO: check if the queue is already present
	_, err = db.Driver.CreateQueue(map[string]interface{}{
		"qid":              qid.String(),
		"name":             name,
		"tasks_waiting":    []string{},
		"tasks_inprogress": []string{},
		"creation_date":    ct,
		"update_date":      ct,
	})

	ctx.APICreationSuccess(qid.String(), "queue")

	return err
}

func Create(queue NodeQueue, ctx *context.Context, db *database.Database) (string, error) {
	// TODO: Add fields check

	ct := time.Now().UTC().Format("20060102150405")

	docID, err := db.Driver.CreateNodeQueues(map[string]interface{}{
		"akey":          queue.AgentKey,
		"nodeid":        queue.NodeId,
		"queues":        queue.Queues,
		"creation_date": ct,
	})

	return docID, err
}

func AddTaskInProgress(ctx *context.Context, db *database.Database) error {
	qid := ctx.Params(":qid")
	tid := ctx.Params(":tid")

	if qid == "" {
		return errors.New("Invalid queue id")
	}

	if tid == "" {
		return errors.New("Invalid task id")
	}

	err := db.Driver.AddTaskInProgress2Queue(qid, tid)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func AddTaskInWaiting(ctx *context.Context, db *database.Database) error {
	qid := ctx.Params(":qid")
	tid := ctx.Params(":tid")

	if qid == "" {
		return errors.New("Invalid queue id")
	}

	if tid == "" {
		return errors.New("Invalid task id")
	}

	err := db.Driver.AddTaskInWaiting2Queue(qid, tid)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

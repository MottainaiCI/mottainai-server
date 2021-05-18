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

type QueueFilter struct {
	Queues []string `json:"queues,omitempty"`
}

func Show(ctx *context.Context, db *database.Database) {

	id := ctx.Params(":id")
	queue, err := db.Driver.GetNodeQueues(id)
	if err != nil {
		ctx.NotFound()
		return
	}
	ctx.JSON(200, queue)

}

func ShowAll(ctx *context.Context, db *database.Database) {
	queues := db.Driver.AllNodesQueues()

	ctx.JSON(200, queues)
}

func ShowAllQueues(filter QueueFilter, ctx *context.Context, db *database.Database) {
	queues := db.Driver.AllQueues(filter.Queues)

	ctx.JSON(200, queues)
}

func GetQid(ctx *context.Context, db *database.Database) error {
	q := ctx.Params(":name")

	if q == "" {
		return errors.New("Invalid queue name")
	}

	queue, err := db.Driver.GetQueueByKey(q)
	if err != nil {
		return err
	}

	ctx.JSON(200, queue.Qid)

	return nil
}

func ShowNode(queue NodeQueue, ctx *context.Context, db *database.Database) error {
	if queue.AgentKey == "" {
		return errors.New("Invalid agent key")
	}

	if queue.NodeId == "" {
		return errors.New("Invalid node id")
	}

	nq, err := db.Driver.GetNodeQueuesByKey(queue.AgentKey, queue.NodeId)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	ctx.JSON(200, nq)

	return nil
}

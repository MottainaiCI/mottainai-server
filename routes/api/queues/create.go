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
	"time"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
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

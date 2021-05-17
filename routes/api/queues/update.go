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

func AddTask(queue NodeQueue, ctx *context.Context, db *database.Database) error {
	q := ctx.Params(":queue")
	tid := ctx.Params(":tid")

	if q == "" {
		return errors.New("Invalid queue")
	}

	if tid == "" {
		return errors.New("Invalid task id")
	}

	err := db.Driver.AddNodeQueuesTask(queue.AgentKey, queue.NodeId, q, tid)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

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
	"fmt"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
)

func Show(ctx *context.Context, db *database.Database) {

	id := ctx.Params(":id")
	queue, err := db.Driver.GetNodeQueues(id)
	fmt.Println("QUEUE ", queue)
	if err != nil {
		ctx.NotFound()
		return
	}
	ctx.JSON(200, queue)

}

func ShowAll(ctx *context.Context, db *database.Database) {
	nodes := db.Driver.AllNodesQueues()

	ctx.JSON(200, nodes)
}

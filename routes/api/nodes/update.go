/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package nodesapi

import (
	"fmt"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/nodes"

	rabbithole "github.com/michaelklishin/rabbit-hole"
)

func Register(nodedata nodes.Node, rmqc *rabbithole.Client, ctx *context.Context, db *database.Database) string {
	key := nodedata.Key
	nodeid := nodedata.NodeID
	hostname := nodedata.Hostname
	fmt.Println("KEY " + key + ", ID " + nodeid)

	if len(key) == 0 {
		ctx.NotFound()
		return ":("
	}

	nodesfound, err := db.FindDoc("Nodes", `[{"eq": "`+key+`", "in": ["key"]}]`)
	if err != nil || len(nodesfound) > 1 || len(nodesfound) == 0 {
		ctx.NotFound()
		return ":("
	}

	//var mynode nodes.Node
	var mynodeid int
	// Query result are document IDs
	for id := range nodesfound {
		//	mynode, _ = db.GetNode(id)
		mynodeid = id
	}

	//	if mynode.NodeID != "" { //Already registered
	//ctx.NotFound()
	//return ":("
	//	}

	db.UpdateNode(mynodeid, map[string]interface{}{
		"nodeid":   nodeid,
		"hostname": hostname,
	})

	return "OK"
}

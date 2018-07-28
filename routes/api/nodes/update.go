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
	"strconv"
	"strings"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	rabbithole "github.com/michaelklishin/rabbit-hole"
)

type NodeUpdate struct {
	NodeID   string `form:"nodeid" json:"nodeid"`
	Key      string `json:"key" form:"key"`
	Hostname string `json:"hostname" form:"hostname"`
}

func Register(nodedata NodeUpdate, rmqc *rabbithole.Client, ctx *context.Context, db *database.Database) string {
	key := nodedata.Key
	nodeid := nodedata.NodeID
	hostname := nodedata.Hostname

	if len(key) == 0 {
		ctx.NotFound()
		return ":("
	}

	n := db.AllNodes()

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
	var pos int
	for p, i := range n {
		if i.ID == mynodeid {
			pos = p
		}
	}

	hb := time.Now().Format("20060102150405")
	db.UpdateNode(mynodeid, map[string]interface{}{
		"nodeid":      nodeid,
		"hostname":    hostname,
		"last_report": hb,
	})

	return strings.Join([]string{strconv.Itoa(len(n)), strconv.Itoa(pos)}, ",")
}

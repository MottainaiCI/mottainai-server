/*

Copyright (C) 2017-2021  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>
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

	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
)

type NodeUpdate struct {
	NodeID     string         `form:"nodeid" json:"nodeid"`
	Key        string         `json:"key" form:"key"`
	Hostname   string         `json:"hostname" form:"hostname"`
	Standalone bool           `json:"standalone" form:"standalone"`
	Queues     map[string]int `json:"queues" form:"queues"`
}

func Register(nodedata NodeUpdate, ctx *context.Context, db *database.Database) error {
	key := nodedata.Key
	nodeid := nodedata.NodeID
	hostname := nodedata.Hostname

	fmt.Println("RECEIVED UPDATE ", nodedata)

	if len(key) == 0 {
		return errors.New("Invalid key")
	}

	n := db.Driver.AllNodes()

	nodefound, err := db.Driver.GetNodeByKey(key)
	if err != nil {
		ctx.NotFound()
		return nil
	}

	hb := time.Now().Format("20060102150405")
	doc := map[string]interface{}{
		"nodeid":      nodeid,
		"hostname":    hostname,
		"last_report": hb,
		"standalone":  nodedata.Standalone,
	}

	if !nodefound.OverrideQueues {
		doc["queues"] = nodedata.Queues
	}

	// Find my position between nodes
	var pos int
	for p, i := range n {
		if i.ID == nodefound.ID {
			pos = p
		}
	}

	db.Driver.UpdateNode(nodefound.ID, doc)

	ctx.APIEventData(strings.Join(
		[]string{strconv.Itoa(len(n)),
			strconv.Itoa(pos)}, ","),
	)
	return nil
}

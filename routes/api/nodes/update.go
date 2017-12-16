package nodesapi

import (
	"fmt"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/nodes"

	machinery "github.com/RichardKnop/machinery/v1"
	rabbithole "github.com/michaelklishin/rabbit-hole"
)

func Register(nodedata nodes.Node, rmqc *rabbithole.Client, ctx *context.Context, rabbit *machinery.Server, db *database.Database) string {
	key := nodedata.Key
	nodeid := nodedata.NodeID
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

	var mynode nodes.Node
	var mynodeid int
	// Query result are document IDs
	for id := range nodesfound {
		mynode, _ = db.GetNode(id)
		mynodeid = id
	}

	if mynode.NodeID != "" { //Already registered
		ctx.NotFound()
		return ":("
	}

	db.UpdateNode(mynodeid, map[string]interface{}{
		"nodeid": nodeid,
	})

	return "OK"
}

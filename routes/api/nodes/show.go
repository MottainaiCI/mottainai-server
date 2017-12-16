package nodesapi

import (
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/db"
)

func Show(ctx *context.Context, db *database.Database) {

	id := ctx.ParamsInt(":id")
	node, err := db.GetNode(id)
	if err != nil {
		ctx.NotFound()
		return
	}
	ctx.JSON(200, node)

}

func ShowAll(ctx *context.Context, db *database.Database) {
	//tasks := db.ListTasks()
	nodes := db.AllNodes()

	ctx.JSON(200, nodes)
}

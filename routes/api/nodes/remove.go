package nodesapi

import (
	"github.com/MottainaiCI/mottainai-server/pkg/context"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	rabbithole "github.com/michaelklishin/rabbit-hole"
)

func APIRemove(rmqc *rabbithole.Client, ctx *context.Context, db *database.Database) string {
	_, err := Remove(rmqc, ctx, db)
	if err != nil {
		ctx.NotFound()
		return ":("
	}
	ctx.Redirect("/nodes")

	return "OK"
}

func Remove(rmqc *rabbithole.Client, ctx *context.Context, db *database.Database) (string, error) {
	id := ctx.ParamsInt(":id")
	node, _ := db.GetNode(id)

	_, err := rmqc.DeleteUser(node.User)
	if err != nil {
		return "", err
	}
	err = db.DeleteNode(id)
	if err != nil {
		return "", err

	}

	return "OK", nil
}

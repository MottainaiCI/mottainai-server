package nodesapi

import (
	"fmt"
	"strconv"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"

	rabbithole "github.com/michaelklishin/rabbit-hole"
)

func APICreate(rmqc *rabbithole.Client, ctx *context.Context, db *database.Database) string {
	id, err := Create(rmqc, ctx, db)
	if err != nil {
		return ":( Error: " + err.Error()
	}

	return id
}

func Create(rmqc *rabbithole.Client, ctx *context.Context, db *database.Database) (string, error) {
	user, _ := utils.RandomString(10)
	pass, _ := utils.RandomString(10)
	key, _ := utils.RandomString(30)

	_, err := rmqc.PutUser(user, rabbithole.UserSettings{Password: pass, Tags: ""})
	if err != nil {
		return "", err
	}
	_, err = rmqc.UpdatePermissionsIn("/", user, rabbithole.Permissions{Configure: ".*", Read: ".*", Write: ".*"})
	if err != nil {
		return "", err
	}

	docID, _ := db.CreateNode(map[string]interface{}{
		"owner": 0,
		"user":  user,
		"pass":  pass,
		"key":   key})
	fmt.Println(docID)

	return strconv.Itoa(docID), nil
}

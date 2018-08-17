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

	docID, _ := db.Driver.CreateNode(map[string]interface{}{
		"owner": 0,
		"user":  user,
		"pass":  pass,
		"key":   key})

	_, err := rmqc.PutUser(user, rabbithole.UserSettings{Password: pass, Tags: ""})
	if err != nil {
		return "", err
	}
	_, err = rmqc.UpdatePermissionsIn("/", user, rabbithole.Permissions{Configure: ".*", Read: ".*", Write: ".*"})
	if err != nil {
		return "", err
	}

	return docID, nil
}

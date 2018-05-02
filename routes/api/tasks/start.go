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

package tasksapi

import (
	"fmt"

	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
)

func APISendStartTask(m *mottainai.Mottainai, th *agenttasks.TaskHandler, ctx *context.Context, db *database.Database) string {
	_, err := SendStartTask(m, th, ctx, db)
	if err != nil {
		ctx.NotFound()
		return ":("
	}
	return "OK"
}

func SendStartTask(m *mottainai.Mottainai, th *agenttasks.TaskHandler, ctx *context.Context, db *database.Database) (string, error) {
	id := ctx.ParamsInt(":id")
	fmt.Println("Starting task ", id)

	mytask, err := db.GetTask(id)
	if err != nil {
		return "", err
	}
	if mytask.IsWaiting() || mytask.IsRunning() {
		return "WAITING/RUNNING", nil
	}

	_, err = m.SendTask(id)
	if err != nil {
		return ":( ", err
	} else {
		return "OK", nil
	}
}

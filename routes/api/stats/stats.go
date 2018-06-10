/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package stats

import (
	context "github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	macaron "gopkg.in/macaron.v1"
)

type Stats struct {
	Running int `json:"running"`
	Waiting int `json:"waiting"`
	Errored int `json:"error"`
	Total   int `json:"total_tasks"`
}

func Info(ctx *context.Context, db *database.Database) {
	rtasks, _ := db.FindDoc("Tasks", `[{"eq": "running", "in": ["status"]}]`)
	running_tasks := len(rtasks)
	wtasks, _ := db.FindDoc("Tasks", `[{"eq": "waiting", "in": ["status"]}]`)
	waiting_tasks := len(wtasks)
	etasks, _ := db.FindDoc("Tasks", `[{"eq": "error", "in": ["result"]}]`)
	error_tasks := len(etasks)

	s := &Stats{}
	total := db.DB().Use("Tasks").ApproxDocCount()
	if total == 0 {
		total = len(db.ListDocs("Tasks"))
	}
	s.Errored = error_tasks
	s.Running = running_tasks
	s.Total = total
	s.Waiting = waiting_tasks

	ctx.JSON(200, s)
}

func Setup(m *macaron.Macaron) {
	m.Get("/api/stats", Info)
}

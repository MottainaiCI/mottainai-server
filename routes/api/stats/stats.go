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
	"strconv"
	"time"

	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	context "github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	macaron "gopkg.in/macaron.v1"
)

type Stats struct {
	Running   int `json:"running"`
	Waiting   int `json:"waiting"`
	Errored   int `json:"error"`
	Failed    int `json:"failed"`
	Succeeded int `json:"succeeded"`
	Total     int `json:"total_tasks"`

	CreatedDaily   map[string]int `json:"created_daily"`
	FailedDaily    map[string]int `json:"failed_daily"`
	ErroredDaily   map[string]int `json:"errored_daily"`
	SucceededDaily map[string]int `json:"succeeded_daily"`
}

func Info(ctx *context.Context, db *database.Database) {
	rtasks, _ := db.Driver.FindDoc("Tasks", `[{"eq": "running", "in": ["status"]}]`)
	running_tasks := len(rtasks)
	wtasks, _ := db.Driver.FindDoc("Tasks", `[{"eq": "waiting", "in": ["status"]}]`)
	waiting_tasks := len(wtasks)
	etasks, _ := db.Driver.FindDoc("Tasks", `[{"eq": "error", "in": ["result"]}]`)
	error_tasks := len(etasks)
	ftasks, _ := db.Driver.FindDoc("Tasks", `[{"eq": "failed", "in": ["result"]}]`)
	failed_tasks := len(ftasks)
	stasks, _ := db.Driver.FindDoc("Tasks", `[{"eq": "success", "in": ["result"]}]`)
	succeeded_tasks := len(stasks)

	var fail_task = make([]agenttasks.Task, 0)
	for i, _ := range ftasks {
		t, _ := db.Driver.GetTask(i)
		fail_task = append(fail_task, t)
	}
	var failed = GetStats(fail_task)

	var err_tasks = make([]agenttasks.Task, 0)
	for i, _ := range etasks {
		t, _ := db.Driver.GetTask(i)
		err_tasks = append(err_tasks, t)
	}
	var errored = GetStats(err_tasks)

	var suc_tasks = make([]agenttasks.Task, 0)
	for i, _ := range stasks {
		t, _ := db.Driver.GetTask(i)
		suc_tasks = append(suc_tasks, t)
	}
	var succeded = GetStats(suc_tasks)

	atasks := db.Driver.AllTasks()
	var created = GetStats(atasks)

	s := &Stats{}
	total := len(atasks)
	s.Errored = error_tasks
	s.Running = running_tasks
	s.Total = total
	s.Waiting = waiting_tasks
	s.Failed = failed_tasks
	s.Succeeded = succeeded_tasks
	s.CreatedDaily = created
	s.FailedDaily = failed
	s.ErroredDaily = errored
	s.SucceededDaily = succeded

	ctx.JSON(200, s)
}

func GetStats(atasks []agenttasks.Task) map[string]int {
	var created = make(map[string]int)
	for _, t := range atasks {
		//t, _ := db.Driver.GetTask(i)
		t1, _ := time.Parse(
			"20060102150405",
			t.CreatedTime)
		day := strconv.Itoa(t1.Year()) + "-" + strconv.Itoa(int(t1.Month())) + "-" + strconv.Itoa(t1.Day())
		i, ok := created[day]
		if !ok {
			created[day] = 1
		} else {
			created[day] = i + 1
		}
	}
	return created
}

func Setup(m *macaron.Macaron) {
	//reqSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: true})

	m.Invoke(func(config *setting.Config) {
		m.Group(config.GetWeb().GroupAppPath(), func() {
			m.Get("/api/stats", Info)
		})
	})
}

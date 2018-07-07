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
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/go-macaron/binding"

	macaron "gopkg.in/macaron.v1"
)

func Setup(m *macaron.Macaron) {
	reqSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: true})

	bind := binding.Bind
	m.Get("/api/tasks", ShowAll)
	m.Post("/api/tasks", reqSignIn, bind(agenttasks.Task{}), APICreate)
	m.Get("/api/tasks/:id", GetTaskJson) // TEMP: For now, as js  calls aren't with auth
	m.Get("/api/tasks/stream_output/:id/:pos", StreamOutputTask)
	m.Get("/api/tasks/tail_output/:id/:pos", TailTask)
	m.Get("/api/tasks/start/:id", reqSignIn, SendStartTask)
	m.Get("/api/tasks/clone/:id", reqSignIn, CloneTask)

	m.Get("/api/tasks/stop/:id", reqSignIn, Stop)
	m.Get("/api/tasks/delete/:id", reqSignIn, APIDelete)
	m.Get("/api/tasks/update", reqSignIn, bind(UpdateTaskForm{}), UpdateTask)
	m.Get("/api/tasks/append", reqSignIn, bind(UpdateTaskForm{}), AppendToTask)
	m.Get("/api/tasks/updatefield", reqSignIn, bind(UpdateTaskForm{}), UpdateTaskField)
	m.Get("/api/tasks/:id/artefacts", reqSignIn, ArtefactList)
	m.Get("/api/artefacts", reqSignIn, AllArtefactList)

	m.Post("/api/tasks/plan", reqSignIn, bind(agenttasks.Plan{}), Plan)
	m.Get("/api/tasks/planned", reqSignIn, PlannedTasks)
	m.Get("/api/tasks/plan/delete/:id", reqSignIn, PlanDelete)
	m.Get("/api/tasks/plan/:id", reqSignIn, PlannedTask)

	m.Post("/api/tasks/artefact/upload", reqSignIn, binding.MultipartForm(ArtefactForm{}), ArtefactUpload)
}

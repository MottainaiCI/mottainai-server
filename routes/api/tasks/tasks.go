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
	"github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	"github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/go-macaron/binding"
)

func ValidateNodeKey(f *UpdateTaskForm, db *database.Database) (bool, int) {
	return ValidateKey(f.APIKey, db)
}

func ValidateKey(k string, db *database.Database) (bool, int) {

	if len(k) == 0 {
		return false, 0
	}

	nodesfound, err := db.FindDoc("Nodes", `[{"eq": "`+k+`", "in": ["key"]}]`)
	if err != nil || len(nodesfound) > 1 || len(nodesfound) == 0 {
		return false, 0
	}

	//var mynode nodes.Node
	var mynodeid = 0
	// Query result are document IDs
	for id := range nodesfound {
		//mynode, _ = db.GetNode(id)
		mynodeid = id
	}
	if mynodeid == 0 {
		return false, 0
	}
	return true, mynodeid
}

func Setup(m *mottainai.Mottainai) {
	bind := binding.Bind
	m.Get("/api/tasks", ShowAll)
	m.Post("/api/tasks", bind(agenttasks.Task{}), APICreate)
	m.Get("/api/tasks/:id", GetTaskJson)
	m.Get("/api/tasks/stream_output/:id/:pos", StreamOutputTask)
	m.Get("/api/tasks/tail_output/:id/:pos", TailTask)
	m.Get("/api/tasks/start/:id", SendStartTask)
	m.Get("/api/tasks/stop/:id", Stop)
	m.Get("/api/tasks/delete/:id", APIDelete)
	m.Get("/api/tasks/update", bind(UpdateTaskForm{}), UpdateTask)
	m.Get("/api/tasks/append", bind(UpdateTaskForm{}), AppendToTask)
	m.Get("/api/tasks/updatefield", bind(UpdateTaskForm{}), UpdateTaskField)
	m.Get("/api/tasks/:id/artefacts", ArtefactList)
	m.Get("/api/artefacts", AllArtefactList)

	m.Post("/api/tasks/artefact/upload", binding.MultipartForm(ArtefactForm{}), ArtefactUpload)
}

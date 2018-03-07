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

package tasks

import (
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	"github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"github.com/go-macaron/binding"
)

func Setup(m *mottainai.Mottainai) {
	bind := binding.BindIgnErr
	m.Get("/tasks", ShowAll)
	m.Get("/tasks/add", Add)
	m.Post("/tasks", bind(agenttasks.Task{}), Create)
	m.Get("/tasks/display/:id", DisplayTask)
	m.Get("/tasks/:id", DisplayTask)
	m.Get("/tasks/start/:id", SendStartTask)
	m.Get("/tasks/stop/:id", Stop)
	m.Get("/tasks/delete/:id", Delete)
}

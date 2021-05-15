/*

Copyright (C) 2021  Daniele Rondina, geaaru@sabayonlinux.org

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

package queuesapi

import (
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	"github.com/go-macaron/binding"
	macaron "gopkg.in/macaron.v1"
)

func Setup(m *macaron.Macaron) {

	m.Invoke(func(config *setting.Config) {
		bind := binding.Bind
		reqSignIn := context.Toggle(&context.ToggleOptions{
			SignInRequired: true,
			Config:         config,
			BaseURL:        config.GetWeb().AppSubURL})
		reqManager := context.Toggle(&context.ToggleOptions{
			ManagerRequired: true,
			Config:          config,
			BaseURL:         config.GetWeb().AppSubURL})

		m.Group(config.GetWeb().GroupAppPath(), func() {

			// NodeQueue actions
			v1.Schema.GetNodeQueueRoute("show_all").ToMacaron(m, reqSignIn, ShowAll)
			v1.Schema.GetNodeQueueRoute("show").ToMacaron(m, reqSignIn, reqManager, Show)
			v1.Schema.GetNodeQueueRoute("create").ToMacaron(
				m, reqSignIn, reqManager, bind(NodeQueue{}), APICreate,
			)
			v1.Schema.GetNodeQueueRoute("add_task").ToMacaron(
				m, reqSignIn, reqManager, bind(NodeQueue{}), AddTask,
			)
			v1.Schema.GetNodeQueueRoute("del_task").ToMacaron(
				m, reqSignIn, reqManager, bind(NodeQueue{}), DelTask,
			)
			v1.Schema.GetNodeQueueRoute("delete").ToMacaron(
				m, reqSignIn, reqManager, bind(NodeQueue{}), Remove,
			)

			// Queue actions
			v1.Schema.GetQueueRoute("show_all").ToMacaron(m, reqSignIn, ShowAllQueues)
			v1.Schema.GetQueueRoute("create").ToMacaron(m, reqSignIn, reqManager, APIQueueCreate)
			v1.Schema.GetQueueRoute("delete").ToMacaron(m, reqSignIn, reqManager, RemoveQueue)
			v1.Schema.GetQueueRoute("add_task_in_progress").ToMacaron(
				m, reqSignIn, reqManager, AddTaskInProgress,
			)
			v1.Schema.GetQueueRoute("del_task_in_progress").ToMacaron(
				m, reqSignIn, reqManager, DelTaskInProgress,
			)
			v1.Schema.GetQueueRoute("add_task").ToMacaron(
				m, reqSignIn, reqManager, AddTaskInWaiting,
			)
			v1.Schema.GetQueueRoute("del_task").ToMacaron(
				m, reqSignIn, reqManager, DelTaskInWaiting,
			)
		})

	})
}

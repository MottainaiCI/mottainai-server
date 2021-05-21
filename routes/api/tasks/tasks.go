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
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	"github.com/go-macaron/binding"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	macaron "gopkg.in/macaron.v1"
)

func Setup(m *macaron.Macaron) {
	m.Invoke(func(config *setting.Config) {
		reqSignIn := context.Toggle(&context.ToggleOptions{
			SignInRequired: true,
			Config:         config,
			BaseURL:        config.GetWeb().AppSubURL})

		bind := binding.Bind
		m.Group(config.GetWeb().GroupAppPath(), func() {
			v1.Schema.GetTaskRoute("show_all").ToMacaron(m, ShowAll)
			v1.Schema.GetTaskRoute("show_all_filtered").ToMacaron(m, ShowAllFiltered)
			v1.Schema.GetTaskRoute("create").ToMacaron(m, reqSignIn, bind(agenttasks.Task{}), APICreate)
			v1.Schema.GetTaskRoute("as_json").ToMacaron(m, GetTaskJson) // TEMP: For now, as js  calls aren't with auth
			v1.Schema.GetTaskRoute("as_yaml").ToMacaron(m, GetTaskYaml) // TEMP: For now, as js  calls aren't with auth
			v1.Schema.GetTaskRoute("stream_output").ToMacaron(m, StreamOutputTask)
			v1.Schema.GetTaskRoute("tail_output").ToMacaron(m, TailTask)
			v1.Schema.GetTaskRoute("start").ToMacaron(m, reqSignIn, SendStartTask)
			v1.Schema.GetTaskRoute("clone").ToMacaron(m, reqSignIn, CloneTask)
			v1.Schema.GetTaskRoute("status").ToMacaron(m, reqSignIn, APIShowTaskByStatus)
			v1.Schema.GetTaskRoute("stop").ToMacaron(m, reqSignIn, APIStop)
			v1.Schema.GetTaskRoute("delete").ToMacaron(m, reqSignIn, APIDelete)
			v1.Schema.GetTaskRoute("update").ToMacaron(m, reqSignIn, bind(UpdateTaskForm{}), UpdateTask)
			v1.Schema.GetTaskRoute("append").ToMacaron(m, reqSignIn, bind(UpdateTaskForm{}), AppendToTask)
			v1.Schema.GetTaskRoute("update_field").ToMacaron(m, reqSignIn, bind(UpdateTaskForm{}), UpdateTaskField)
			v1.Schema.GetTaskRoute("update_node").ToMacaron(m, reqSignIn, bind(UpdateTaskForm{}), SetNode)
			v1.Schema.GetTaskRoute("artefact_list").ToMacaron(m, reqSignIn, ArtefactList)
			v1.Schema.GetTaskRoute("all_artefact_list").ToMacaron(m, reqSignIn, AllArtefactList)
			v1.Schema.GetTaskRoute("artefact_upload").ToMacaron(m, reqSignIn, binding.MultipartForm(ArtefactForm{}), ArtefactUpload)

			v1.Schema.GetTaskRoute("create_plan").ToMacaron(m, reqSignIn, bind(agenttasks.Plan{}), Plan)
			v1.Schema.GetTaskRoute("plan_list").ToMacaron(m, reqSignIn, PlannedTasks)
			v1.Schema.GetTaskRoute("plan_delete").ToMacaron(m, reqSignIn, PlanDelete)
			v1.Schema.GetTaskRoute("plan_show").ToMacaron(m, reqSignIn, PlannedTask)

			v1.Schema.GetTaskRoute("create_pipeline").ToMacaron(m, reqSignIn, bind(agenttasks.PipelineForm{}), Pipeline)
			v1.Schema.GetTaskRoute("pipeline_list").ToMacaron(m, reqSignIn, ShowAllPipelines)
			v1.Schema.GetTaskRoute("pipeline_delete").ToMacaron(m, reqSignIn, PipelineDelete)
			v1.Schema.GetTaskRoute("pipeline_show").ToMacaron(m, reqSignIn, APIPipelineShow)
			v1.Schema.GetTaskRoute("pipeline_as_yaml").ToMacaron(m, reqSignIn, PipelineYaml)
			v1.Schema.GetTaskRoute("pipeline_completed").ToMacaron(m, reqSignIn, PipelineCompleted)
		})
	})
}

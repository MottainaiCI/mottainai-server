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

package apiwebhook

import (
	"errors"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	"github.com/go-macaron/binding"

	macaron "gopkg.in/macaron.v1"
)

func RequiresWebHookSetting(c *context.Context, db *database.Database) error {
	// Check setting if we have to process this.
	err := errors.New("Webhook integration disabled")
	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_ENABLED)
	if err == nil {
		if uuu.IsDisabled() {
			c.ServerError("Webhook integration disabled", err)
			return err
		}
	}
	return nil
}

func Setup(m *macaron.Macaron) {
	m.Invoke(func(config *setting.Config) {
		reqSignIn := context.Toggle(&context.ToggleOptions{
			SignInRequired: true,
			Config:         config,
			BaseURL:        config.GetWeb().AppSubURL,
		})
		bind := binding.Bind

		m.Group(config.GetWeb().GroupAppPath(), func() {
			v1.Schema.GetWebHookRoute("show_all").ToMacaron(m, RequiresWebHookSetting, reqSignIn, ShowAll)
			v1.Schema.GetWebHookRoute("create").ToMacaron(m, RequiresWebHookSetting, reqSignIn, Create)
			v1.Schema.GetWebHookRoute("show").ToMacaron(m, RequiresWebHookSetting, reqSignIn, ShowSingle)
			v1.Schema.GetWebHookRoute("delete").ToMacaron(m, RequiresWebHookSetting, reqSignIn, Remove)
			v1.Schema.GetWebHookRoute("update_task").ToMacaron(m, RequiresWebHookSetting, reqSignIn, bind(agenttasks.Task{}), UpdateTask)
			v1.Schema.GetWebHookRoute("update_pipeline").ToMacaron(m, RequiresWebHookSetting, reqSignIn, bind(agenttasks.PipelineForm{}), UpdatePipeline)
			v1.Schema.GetWebHookRoute("delete_task").ToMacaron(m, RequiresWebHookSetting, reqSignIn, DeleteTask)
			v1.Schema.GetWebHookRoute("delete_pipeline").ToMacaron(m, RequiresWebHookSetting, reqSignIn, DeletePipeline)
			v1.Schema.GetWebHookRoute("set_field").ToMacaron(m, RequiresWebHookSetting, reqSignIn, bind(WebhookUpdate{}), SetWebHookField)
		})
	})
}

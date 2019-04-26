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

package storagesapi

import (
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/go-macaron/binding"
	macaron "gopkg.in/macaron.v1"
)

func Setup(m *macaron.Macaron) {
	m.Invoke(func(config *setting.Config) {
		reqSignIn := context.Toggle(&context.ToggleOptions{
			SignInRequired: true,
			Config:         config,
			BaseURL:        config.GetWeb().AppSubURL})

		m.Group(config.GetWeb().GroupAppPath(), func() {
			v1.Schema.GetStorageRoute("show_all").ToMacaron(m, reqSignIn, StorageList)
			v1.Schema.GetStorageRoute("show_artefacts").ToMacaron(m, reqSignIn, StorageListArtefacts)
			v1.Schema.GetStorageRoute("create").ToMacaron(m, reqSignIn, StorageCreate)
			v1.Schema.GetStorageRoute("delete").ToMacaron(m, reqSignIn, StorageDelete)
			v1.Schema.GetStorageRoute("remove_path").ToMacaron(m, reqSignIn, StorageRemovePath)
			v1.Schema.GetStorageRoute("show").ToMacaron(m, reqSignIn, StorageShow)
			v1.Schema.GetStorageRoute("upload").ToMacaron(m, reqSignIn, binding.MultipartForm(StorageForm{}), StorageUpload)
		})
	})
}

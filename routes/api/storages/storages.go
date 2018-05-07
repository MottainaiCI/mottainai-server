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
	"github.com/go-macaron/binding"
	macaron "gopkg.in/macaron.v1"
)

func Setup(m *macaron.Macaron) {
	//bind := binding.Bind
	m.Get("/api/storage/list", StorageList)
	m.Get("/api/storage/:id/list", StorageListArtefacts)

	m.Get("/api/storage/:name/create", StorageCreate)
	m.Get("/api/storage/:id/delete", StorageDelete)
	m.Get("/api/storage/:id/show", StorageShow)

	m.Post("/api/storage/upload", binding.MultipartForm(StorageForm{}), StorageUpload)

}

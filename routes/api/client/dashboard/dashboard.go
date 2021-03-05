/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
Copyright (C) 2020       Adib Saad <adib.saad@gmail.com>
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

package dashboard

import (
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	"gopkg.in/macaron.v1"
)

func Stats(ctx *macaron.Context, db *database.Database) error {
	statuses, e := db.Driver.GetTaskMetrics()
	if e != nil {
		return e
	}
	ctx.JSON(200, &statuses)
	return nil
}

func Setup(m *macaron.Macaron) {
	v1.Schema.GetClientRoute("dashboard_stats").ToMacaron(m, Stats)
}

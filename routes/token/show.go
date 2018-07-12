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

package tokenroute

import (
	apitoken "github.com/MottainaiCI/mottainai-server/routes/api/token"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/template"
)

func ShowAll(ctx *context.Context, db *database.Database) {
	all, mine, err := apitoken.GetTokens(ctx, db)
	if err != nil {
		ctx.ServerError("Failed finding token", err)
		return
	}

	ctx.Data["AllTokens"] = all
	ctx.Data["UserTokens"] = mine
	template.TemplatePreview(ctx, "tokens")
}

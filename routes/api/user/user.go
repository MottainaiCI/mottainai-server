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

package userapi

import (
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	"github.com/go-macaron/binding"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	macaron "gopkg.in/macaron.v1"
)

func UpdateUser(opts user.UserForm, ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")
	u, err := db.Driver.GetUser(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	if len(opts.Email) > 0 {
		u.Email = opts.Email
	}
	if len(opts.Name) > 0 {
		u.Name = opts.Name
	}
	if len(opts.Password) > 0 {
		u.Password = opts.Password
	}

	err = db.Driver.UpdateUser(id, u.ToMap())
	if err != nil {
		ctx.NotFound()
		return err
	}

	return nil
}

func SetManager(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	u, err := db.Driver.GetUser(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	u.MakeManager()

	err = db.Driver.UpdateUser(id, u.ToMap())
	if err != nil {
		ctx.NotFound()
		return err
	}

	return nil
}

func SetAdmin(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	u, err := db.Driver.GetUser(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	u.MakeAdmin()

	err = db.Driver.UpdateUser(id, u.ToMap())
	if err != nil {
		ctx.NotFound()
		return err
	}

	return nil
}

func SetManagerUser(ctx *context.Context, db *database.Database) error {
	err := SetManager(ctx, db)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func SetAdminUser(ctx *context.Context, db *database.Database) error {
	err := SetAdmin(ctx, db)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func UnSetManager(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	u, err := db.Driver.GetUser(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	u.RemoveManager()

	err = db.Driver.UpdateUser(id, u.ToMap())
	if err != nil {
		ctx.NotFound()
		return err
	}
	return nil
}

func UnSetAdmin(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	u, err := db.Driver.GetUser(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	u.RemoveAdmin()

	err = db.Driver.UpdateUser(id, u.ToMap())
	if err != nil {
		ctx.NotFound()
		return err
	}
	return nil
}

func UnSetAdminUser(ctx *context.Context, db *database.Database) error {
	err := UnSetAdmin(ctx, db)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func UnSetManagerUser(ctx *context.Context, db *database.Database) error {
	err := UnSetManager(ctx, db)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func Delete(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	user, err := db.Driver.GetUser(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	tokens, _ := db.Driver.GetTokensByUserID(user.ID)

	for _, t := range tokens {
		db.Driver.DeleteToken(t.ID)
	}

	err = db.Driver.DeleteUser(id)
	if err != nil {
		ctx.NotFound()
		return err
	}
	return nil
}

func DeleteUser(ctx *context.Context, db *database.Database) error {

	err := Delete(ctx, db)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

func List(c *context.Context, db *database.Database) []user.User {
	us := db.Driver.AllUsers()
	return us
}

func Show(c *context.Context, db *database.Database) (user.User, error) {
	id := c.Params(":id")
	u, err := db.Driver.GetUser(id)
	u.Password = ""
	if err != nil {
		return u, nil
	}
	return u, nil
}

func ShowUser(c *context.Context, db *database.Database) {
	u, err := Show(c, db)

	if err != nil {
		c.NotFound()
		return
	}
	c.JSON(200, u)
}

func ListUsers(c *context.Context, db *database.Database) {
	c.JSON(200, List(c, db))
}

func CreateUser(ctx *context.Context, db *database.Database, opts user.UserForm) error {
	var u *user.User = &user.User{
		Name:     opts.Name,
		Email:    opts.Email,
		Password: opts.Password,
		// For now create only normal user. Upgrade to admin/manager
		// is done through specific api call.
		Admin:   "no",
		Manager: "no",
	}

	r, err := db.Driver.InsertAndSaltUser(u)
	if err != nil {
		return err
	}

	ctx.APICreationSuccess(r, "user")
	return nil
}

func Setup(m *macaron.Macaron) {
	m.Invoke(func(config *setting.Config) {
		bind := binding.BindIgnErr
		reqSignIn := context.Toggle(&context.ToggleOptions{
			SignInRequired: true,
			Config:         config,
			BaseURL:        config.GetWeb().AppSubURL})
		reqAdmin := context.Toggle(&context.ToggleOptions{
			AdminRequired: true,
			Config:        config,
			BaseURL:       config.GetWeb().AppSubURL})
		reqManager := context.Toggle(&context.ToggleOptions{
			ManagerRequired: true,
			Config:          config,
			BaseURL:         config.GetWeb().AppSubURL})

		m.Group(config.GetWeb().GroupAppPath(), func() {
			v1.Schema.GetUserRoute("show_all").ToMacaron(m, reqManager, reqSignIn, ListUsers)
			v1.Schema.GetUserRoute("show").ToMacaron(m, reqSignIn, ShowUser)
			v1.Schema.GetUserRoute("set_admin").ToMacaron(m, reqSignIn, reqAdmin, SetAdminUser)
			v1.Schema.GetUserRoute("unset_admin").ToMacaron(m, reqSignIn, reqAdmin, UnSetAdminUser)

			v1.Schema.GetUserRoute("set_manager").ToMacaron(m, reqSignIn, reqAdmin, SetManagerUser)
			v1.Schema.GetUserRoute("unset_manager").ToMacaron(m, reqSignIn, reqAdmin, UnSetManagerUser)
			v1.Schema.GetUserRoute("delete").ToMacaron(m, reqSignIn, reqAdmin, DeleteUser)
			v1.Schema.GetUserRoute("create").ToMacaron(m, reqSignIn, reqAdmin, bind(user.UserForm{}), CreateUser)
			v1.Schema.GetUserRoute("edit").ToMacaron(m, reqSignIn, reqAdmin, bind(user.UserForm{}), UpdateUser)
		})
	})
}

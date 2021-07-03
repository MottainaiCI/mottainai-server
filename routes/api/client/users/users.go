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

package users

import (
	"errors"
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"gopkg.in/macaron.v1"

	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	"github.com/go-macaron/binding"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type UserForm struct {
	Email    string `binding:"Required;Email;MaxSize(254)" form:"email"`
	Name     string `binding:"Required;AlphaDashDot;MaxSize(35)" form:"name"`
	Password string `binding:"Required;MaxSize(255)" form:"password"`
	Admin    bool   `form:"is_admin"`
	Manager  bool   `form:"is_manager"`
}

func UpdateUser(opts UserForm, ctx *context.Context, db *database.Database) {
	id := ctx.Params(":id")
	u, err := db.Driver.GetUser(id)
	if err != nil {
		ctx.NotFound()
		return
	}

	if len(opts.Email) > 0 {
		u.Email = opts.Email
	}
	if len(opts.Name) > 0 {
		u.Name = opts.Name
	}
	if len(opts.Password) > 0 {
		u.Password = opts.Password
		u.SaltPassword()
	}

	if opts.Admin {
		u.Admin = "yes"
	} else {
		u.Admin = "no"
	}

	if opts.Manager {
		u.Manager = "yes"
	} else {
		u.Manager = "no"
	}

	err = db.Driver.UpdateUser(id, u.ToMap())
	if err != nil {
		ctx.NotFound()
		return
	}

	ctx.APIActionSuccess()
}

func Delete(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	if ctx.User.ID == id {
		return errors.New("can't delete yourself")
	}

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

func DeleteUser(ctx *context.Context, db *database.Database) {
	err := Delete(ctx, db)
	if err != nil {
		ctx.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	ctx.APIActionSuccess()
}

func Show(ctx *context.Context, db *database.Database) (user.User, error) {
	id := ctx.Params(":id")
	u, err := db.Driver.GetUser(id)
	u.Password = ""
	if err != nil {
		return u, err
	}
	return u, nil
}

func ShowUser(ctx *context.Context, db *database.Database) {
	u, err := Show(ctx, db)

	if err != nil {
		ctx.NotFound()
		return
	}
	ctx.JSON(200, u)
}

func ListUsers(ctx *context.Context, db *database.Database) {
	us := db.Driver.AllUsers()
	ctx.JSON(200, us)
}

func CreateUser(ctx *context.Context, db *database.Database, opts UserForm) {
	admin := "no"
	manager := "no"
	if opts.Admin {
		admin = "yes"
	}
	if opts.Manager {
		manager = "yes"
	}
	var u *user.User = &user.User{
		Name:     opts.Name,
		Email:    opts.Email,
		Password: opts.Password,
		// For now create only normal user. Upgrade to admin/manager
		// is done through specific api call.
		Admin:   admin,
		Manager: manager,
	}

	r, err := db.Driver.InsertAndSaltUser(u)
	if err != nil {
		ctx.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	ctx.APICreationSuccess(r, "user")
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
			v1.Schema.GetClientRoute("users_show_all").ToMacaron(m, reqManager, reqSignIn, ListUsers)
			v1.Schema.GetClientRoute("users_show").ToMacaron(m, reqSignIn, ShowUser)
			v1.Schema.GetClientRoute("users_create").ToMacaron(m, reqSignIn, reqAdmin, bind(UserForm{}), CreateUser)
			v1.Schema.GetClientRoute("users_edit").ToMacaron(m, reqSignIn, reqAdmin, bind(UserForm{}), UpdateUser)
			v1.Schema.GetClientRoute("users_delete").ToMacaron(m, reqSignIn, reqAdmin, DeleteUser)
		})
	})
}

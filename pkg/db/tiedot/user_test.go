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

package tiedot

import (
	"os"
	"testing"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"
)

var dbtest4 *Database

func TestInsertUser(t *testing.T) {

	config := setting.NewConfig(nil)
	// Set env variable
	config.Viper.SetEnvPrefix(setting.MOTTAINAI_ENV_PREFIX)
	config.Viper.AutomaticEnv()
	config.Viper.SetTypeByDefaultValue(true)
	config.Unmarshal()

	config.GetDatabase().DBPath = "./DB"
	db := New(config.GetDatabase().DBPath)
	db.GetAgent().Map(config)
	db.Init()
	dbtest4 = db
	u := &user.User{}
	u.Name = "test"
	u.Password = "foo"
	u.Email = "foo@bar"
	id, err := db.InsertAndSaltUser(u)

	if err != nil {
		t.Fatal("Failed insert")
	}

	uu, _ := db.GetUser(id)

	if uu.Name != u.Name {
		t.Fatal("Failed insert")
	}

	_, err = db.InsertAndSaltUser(u)
	if err == nil {
		t.Fatal("User could have been created twice")
	}

	db.DeleteUser(id)

	_, err = db.GetUser(id)

	if err == nil {
		t.Fatal("Failed Remove")
	}

}

func TestGetUserByName(t *testing.T) {
	db := dbtest4

	u := &user.User{}
	u.Name = "test2"
	u.Password = "foo"
	u.Email = "foo@bar"
	id, err := db.InsertAndSaltUser(u)

	if err != nil {
		t.Fatal("Failed insert", err)
	}

	uu, _ := db.GetUser(id)

	if uu.Name != u.Name {
		t.Fatal("Failed insert (name differs)")
	}

	uuu, err := db.GetUserByName("test2")

	if err != nil {
		t.Fatal(err)
	}
	if uuu.Name != "test2" {
		t.Fatal("Could not find the inserted user")
	}

	uuu.AddIdentity("github", &user.Identity{ID: "foo"})
	err = db.UpdateUser(id, uuu.ToMap())
	if err != nil {
		t.Fatal(err)
	}

	users, err := db.GetUserByIdentity("github", "foo")
	if err != nil {
		t.Fatal(err)
	}
	if users.Identities["github"].ID != "foo" {
		t.Fatal("Failed decoding identities")
	}
	uu, _ = db.GetUser(id)
	if uu.Identities["github"].ID != "foo" {
		t.Fatal("Failed decoding identities")
	}

	uu.RemoveIdentity("github")

	err = db.UpdateUser(id, uu.ToMap())
	if err != nil {
		t.Fatal(err)
	}

	uu, _ = db.GetUser(id)
	if _, ok := uu.Identities["github"]; ok {
		t.Fatal("Failed removing identities")
	}

	users, err = db.GetUserByIdentity("github", "foo")
	if err == nil {
		t.Fatal("Found identity even if removed")
	}

}

func TestLogin(t *testing.T) {
	var dbpath string
	db := dbtest4

	db.Invoke(func(config *setting.Config) {
		dbpath = config.GetDatabase().DBPath
	})
	defer os.RemoveAll(dbpath)

	u, err := db.SignIn("test2", "foo")

	if err != nil {
		t.Fatal("Failed login", err)
	}
	if u.Name != "test2" {
		t.Fatal("Could not find the inserted user")
	}

	if count := db.CountUsers(); count != 1 {
		t.Fatal("DB Count is (expected 1)", count)
	}

	db.DeleteUser(u.ID)

}

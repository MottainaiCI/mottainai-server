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

package database

import (
	"os"
	"testing"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"
)

func TestInsertUser(t *testing.T) {

	setting.Configuration.DBPath = "./DB"
	db := NewDatabase("")
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

	db.DeleteUser(id)

	_, err = db.GetUser(id)

	if err == nil {
		t.Fatal("Failed Remove")
	}

}

func TestGetUserByName(t *testing.T) {
	db := Instance()

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

}

func TestLogin(t *testing.T) {
	db := Instance()

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

	defer os.RemoveAll(setting.Configuration.DBPath)
}

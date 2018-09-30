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
	token "github.com/MottainaiCI/mottainai-server/pkg/token"
)

var dbtest3 *Database

func TestInsertToken(t *testing.T) {

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
	dbtest3 = db
	u := &token.Token{}
	u.Key = "test"

	id, err := db.InsertToken(u)

	if err != nil {
		t.Fatal("Failed insert")
	}

	uu, _ := db.GetToken(id)

	if uu.Key != u.Key {
		t.Fatal("Failed insert")
	}

	db.DeleteToken(id)

	err = db.DeleteToken(id)

	if err == nil {
		t.Fatal("Failed Remove")
	}

}

func TestGetTokenByKey(t *testing.T) {

	db := dbtest3

	u := &token.Token{}
	u.Key = "test2"
	id, err := db.InsertToken(u)

	if err != nil {
		t.Fatal("Failed insert", err)
	}

	uu, _ := db.GetToken(id)

	if uu.Key != u.Key {
		t.Fatal("Failed insert (Key differs)")
	}

	uuu, err := db.GetTokenByKey("test2")

	if err != nil {
		t.Fatal(err)
	}
	if uuu.Key != "test2" {
		t.Fatal("Could not find the inserted token")
	}

}

func TestGetTokenByUid(t *testing.T) {
	var dbpath string
	db := dbtest3
	db.GetAgent().Invoke(func(config *setting.Config) {
		dbpath = config.GetDatabase().DBPath
	})
	defer os.RemoveAll(dbpath)

	u := &token.Token{}
	u.Key = "test2"
	u.UserId = "20"
	id, err := db.InsertToken(u)

	if err != nil {
		t.Fatal("Failed insert", err)
	}

	uu, _ := db.GetToken(id)

	if uu.Key != u.Key {
		t.Fatal("Failed insert (Key differs)")
	}

	uuu, err := db.GetTokenByKey("test2")

	if err != nil {
		t.Fatal(err)
	}
	if uuu.Key != "test2" {
		t.Fatal("Could not find the inserted token")
	}

	uuuu, err := db.GetTokenByUserID("20")

	if err != nil {
		t.Fatal(err)
	}
	if uuuu.Key != "test2" {
		t.Fatal("Could not find the inserted token")
	}

}

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

	organization "github.com/MottainaiCI/mottainai-server/pkg/organization"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

var DB *Database

func TestInsertOrganization(t *testing.T) {

	config := setting.NewConfig(nil)
	// Set env variable
	config.Viper.SetEnvPrefix(setting.MOTTAINAI_ENV_PREFIX)
	config.Viper.AutomaticEnv()
	config.Viper.SetTypeByDefaultValue(true)
	config.Unmarshal()

	config.GetDatabase().DBPath = "./DB"
	defer os.RemoveAll(config.GetDatabase().DBPath)

	db := New(config.GetDatabase().DBPath)
	db.GetAgent().Map(config)
	db.Init()
	DB = db
	u := &organization.Organization{}
	u.Name = "test"
	u.AddMember("fakemember")
	u.AddOwner("fakeowner")
	id, err := db.InsertOrganization(u)

	if err != nil {
		t.Fatal("Failed insert")
	}

	uu, _ := db.GetOrganization(id)

	if uu.Name != u.Name {
		t.Fatal("Failed insert", uu)
	}

	if !uu.ContainsOwner("fakeowner") {
		t.Fatal("Failed insert", uu)
	}

	if !uu.ContainsMember("fakemember") {
		t.Fatal("Failed insert", uu)
	}

	err = db.DeleteOrganization(id)

	if err != nil {
		t.Fatal("Failed Remove")
	}

}

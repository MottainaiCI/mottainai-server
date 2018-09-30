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
)

var dbtestt *Database

func TestInsertSetting(t *testing.T) {

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
	dbtestt = db
	u := &setting.Setting{}
	u.Key = "test"
	u.Value = "foo"

	id, err := db.InsertSetting(u)

	if err != nil {
		t.Fatal("Failed insert")
	}

	uu, _ := db.GetSetting(id)

	if uu.Key != u.Key {
		t.Fatal("Failed insert")
	}

	if uu.Value != u.Value {
		t.Fatal("Failed insert")
	}

	db.DeleteSetting(id)

	err = db.DeleteSetting(id)

	if err == nil {
		t.Fatal("Failed Remove")
	}

}

func TestGetSettingByKey(t *testing.T) {

	db := dbtestt

	u := &setting.Setting{}
	u.Key = "test2"
	u.Value = "bar"
	id, err := db.InsertSetting(u)

	if err != nil {
		t.Fatal("Failed insert", err)
	}

	uu, _ := db.GetSetting(id)

	if uu.Key != u.Key {
		t.Fatal("Failed insert (Key differs)")
	}

	if uu.Value != u.Value {
		t.Fatal("Failed insert (Key differs)")
	}

	uuu, err := db.GetSettingByKey("test2")

	if err != nil {
		t.Fatal(err)
	}
	if uuu.Key != "test2" {
		t.Fatal("Could not find the inserted setting")
	}
	if uuu.Value != "bar" {
		t.Fatal("Could not find the inserted setting")
	}

	err = db.DeleteSetting(uuu.ID)
	if err != nil {
		t.Fatal("Failed Remove")
	}

}

func TestGetSettingByUid(t *testing.T) {

	var dbpath string
	db := dbtestt

	db.Invoke(func(config *setting.Config) {
		dbpath = config.GetDatabase().DBPath
	})
	defer os.RemoveAll(dbpath)

	u := &setting.Setting{}
	u.Key = "test2"
	u.Value = "20"
	id, err := db.InsertSetting(u)

	if err != nil {
		t.Fatal("Failed insert", err)
	}

	_, err = db.InsertSetting(u)
	if err == nil {
		t.Fatal("Cannot insert same setting twice", err)
	}

	uu, _ := db.GetSetting(id)

	if uu.Key != u.Key {
		t.Fatal("Failed insert (Key differs)")
	}

	uuu, err := db.GetSettingByKey("test2")

	if err != nil {
		t.Fatal(err)
	}
	if uuu.Key != "test2" {
		t.Fatal("Could not find the inserted setting")
	}

	if uuu.Value != "20" {
		t.Fatal("Could not find the inserted setting")
	}
}

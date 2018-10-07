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
	"strconv"
	"testing"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

var dbtest *Database

func TestInsertTask(t *testing.T) {

	config := setting.NewConfig(nil)
	// Set env variable
	config.Viper.SetEnvPrefix(setting.MOTTAINAI_ENV_PREFIX)
	config.Viper.AutomaticEnv()
	config.Viper.SetTypeByDefaultValue(true)
	config.Unmarshal()

	config.Database.DBPath = "./DB"
	db := New(config.GetDatabase().DBPath)
	db.GetAgent().Map(config)
	db.Init()
	dbtest = db
	u := &task.Task{}
	u.Namespace = "docker_execute"
	u.Owner = "20"

	id, err := db.InsertTask(u)

	if err != nil {
		t.Fatal("Failed insert")
	}

	uu, _ := db.GetTask(config, id)

	if uu.Namespace != u.Namespace {
		t.Fatal("Failed insert")
	}
	if uu.Node != u.Node {
		t.Fatal("Failed insert")
	}

	tasks, err := db.AllUserTask(config, "20")

	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fatal("Failed search", tasks)
	}

	if tasks[0].Namespace != u.Namespace {
		t.Fatal("Failed insert")
	}

	db.DeleteTask(config, id)

	err = db.DeleteTask(config, id)

	if err == nil {
		t.Fatal("Failed Remove")
	}

}

func TestUpdateTask(t *testing.T) {
	var dbpath string
	var conf *setting.Config
	db := dbtest
	db.GetAgent().Invoke(func(config *setting.Config) {
		dbpath = config.GetDatabase().DBPath
		conf = config
	})
	defer os.RemoveAll(dbpath)

	u := &task.Task{}
	u.Namespace = "docker_execute"
	//u.Node = "20"

	id, err := db.InsertTask(u)

	if err != nil {
		t.Fatal("Failed insert")
	}

	db.UpdateTask(id, map[string]interface{}{
		"owner_id": strconv.Itoa(20),
	})
	db.UpdateTask(id, map[string]interface{}{
		"node_id": strconv.Itoa(50),
	})
	uu, _ := db.GetTask(conf, id)

	if uu.Namespace != u.Namespace {
		t.Fatal("Failed insert")
	}
	if uu.Owner != strconv.Itoa(20) {
		t.Fatal("Failed insert")
	}

	tasks, err := db.AllUserTask(conf, "20")

	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fatal("Failed search")
	}

	if tasks[0].Namespace != u.Namespace {
		t.Fatal("Failed insert")
	}

	tasks, err = db.AllNodeTask(conf, "50")

	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fatal("Failed search")
	}

	if tasks[0].Namespace != u.Namespace {
		t.Fatal("Failed insert")
	}

	db.DeleteTask(conf, id)

	err = db.DeleteTask(conf, id)

	if err == nil {
		t.Fatal("Failed Remove")
	}

}

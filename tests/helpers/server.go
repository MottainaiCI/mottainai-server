/*

Copyright (C) 2019  Ettore Di Giacinto <mudler@gentoo.org>

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
package tests_helpers

import (
	"github.com/pkg/errors"

	"os"
	"path/filepath"
	"time"

	token "github.com/MottainaiCI/mottainai-server/pkg/token"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	node "github.com/MottainaiCI/mottainai-server/pkg/nodes"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"

	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/routes"
	"github.com/go-macaron/binding"
)

var Config *setting.Config
var Tokens []*token.Token
var Nodes []*node.Node
var Tasks []string

func InitConfig(path string) {
	//os.RemoveAll(path)
	Config = setting.NewConfig(nil)
	// Set env variable
	Config.Viper.SetEnvPrefix(setting.MOTTAINAI_ENV_PREFIX)
	Config.Viper.AutomaticEnv()
	Config.Viper.SetDefault("config", "")
	Config.Viper.SetDefault("etcd-config", false)
	Config.Viper.SetTypeByDefaultValue(true)
	Config.Unmarshal()

	dbpath := filepath.Join(path, "db")
	artefactspath := filepath.Join(path, "artefacts")
	namespacespath := filepath.Join(path, "namespaces")
	storagepath := filepath.Join(path, "sorage")
	lockspath := filepath.Join(path, "locks")

	os.MkdirAll(artefactspath, os.ModePerm)
	os.MkdirAll(dbpath, os.ModePerm)
	os.MkdirAll(namespacespath, os.ModePerm)
	os.MkdirAll(storagepath, os.ModePerm)
	os.MkdirAll(lockspath, os.ModePerm)

	Config.GetWeb().HTTPAddr = "0.0.0.0"
	Config.GetWeb().HTTPPort = "9020"
	Config.GetWeb().AppURL = "http://127.0.0.1:9020"
	Config.GetWeb().LockPath = lockspath
	Config.GetDatabase().DBPath = dbpath
	Config.GetStorage().ArtefactPath = artefactspath
	Config.GetStorage().NamespacePath = namespacespath
	Config.GetStorage().StoragePath = storagepath
}

func SetFixture(db *database.Database) error {

	u := &user.User{}
	u.Name = "test"
	u.Password = "foo"
	u.Email = "foo@bar"
	u.MakeAdmin()
	id, err := db.Driver.InsertAndSaltUser(u)
	if err != nil {
		return errors.Wrap(err, "Error inserting the fixture user")
	}

	tok, err := token.GenerateUserToken(id)
	if err != nil {
		return errors.Wrap(err, "Error generating the fixture token")
	}

	node := &node.Node{Key: "test"}
	_, err = db.Driver.InsertNode(node)
	if err != nil {
		return errors.Wrap(err, "Error inserting the fixture node")
	}
	_, err = db.Driver.InsertToken(tok)
	if err != nil {
		return errors.Wrap(err, "Error inserting the fixture token")
	}

	Tokens = append(Tokens, tok)
	Nodes = append(Nodes, node)

	return nil
}

// Dup for other test suites
func NewClient() (*client.Fetcher, error) {
	if len(Tokens) == 0 {
		return nil, errors.New("No tokens registered in the helper")
	}

	return client.NewTokenClient(Config.GetWeb().AppURL, Tokens[0].Key, Config), nil
}

func SetRuntimeFixture() error {

	c, err := NewClient()
	if err != nil {
		return err
	}

	dat := make(map[string]interface{})

	var flagsName []string = []string{
		"script", "storage", "source", "directory", "task", "image",
		"namespace", "storage_path", "artefact_path", "tag_namespace",
		"prune", "queue", "cache_image",
	}

	for _, n := range flagsName {
		dat[n] = "test"
	}

	res, err := c.GenericForm("/api/tasks", dat)
	if err != nil {
		return errors.Wrap(err, "Failed to create task fixture")
	}
	tid := string(res)
	if tid == "0" {
		return errors.New("Document not created")
	}

	Tasks = append(Tasks, tid)

	return nil
}

func StartServer(path string) error {
	InitConfig(path)
	binding.MaxMemory = int64(1024 * 1024 * 1)

	defer os.RemoveAll(Config.GetDatabase().DBPath)
	defer os.RemoveAll(Config.GetStorage().ArtefactPath)
	db := InitDB(Config)
	defer CleanDB()

	err := SetFixture(db)
	if err != nil {
		return err
	}
	server := mottainai.Classic(Config)
	routes.SetupWebUI(server)
	go server.Start()
	time.Sleep(time.Duration(5 * time.Second))
	err = SetRuntimeFixture()
	if err != nil {
		return errors.Wrap(err, "While creating the runtime fixtures")
	}
	return nil
}

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
	"fmt"

	"github.com/MottainaiCI/mottainai-server/pkg/artefact"
	"github.com/MottainaiCI/mottainai-server/pkg/namespace"
	"github.com/MottainaiCI/mottainai-server/pkg/nodes"
	organization "github.com/MottainaiCI/mottainai-server/pkg/organization"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/storage"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	token "github.com/MottainaiCI/mottainai-server/pkg/token"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
	tiedot "github.com/MottainaiCI/mottainai-server/pkg/db/tiedot"
)

type DatabaseDriver interface {
	Init()
	InsertDoc(string, map[string]interface{}) (int, error)
	FindDoc(string, string) (map[int]struct{}, error)
	DeleteDoc(string, int) error
	UpdateDoc(string, int, map[string]interface{}) error
	ReplaceDoc(string, int, map[string]interface{}) error
	GetDoc(string, int) (map[string]interface{}, error)
	DropColl(string) error
	ListDocs(string) []dbcommon.DocItem
	// Artefacts
	IndexArtefacts()
	CreateArtefact(map[string]interface{}) (int, error)
	DeleteArtefact(int) error
	UpdateArtefact(int, map[string]interface{}) error
	GetArtefact(int) (artefact.Artefact, error)
	SearchArtefact(string) (artefact.Artefact, error)
	ListArtefacts() []dbcommon.DocItem
	AllArtefacts() []artefact.Artefact

	// Namespaces
	IndexNamespace()
	CreateNamespace(t map[string]interface{}) (int, error)
	DeleteNamespace(docID int) error
	UpdateNamespace(docID int, t map[string]interface{}) error
	SearchNamespace(name string) (namespace.Namespace, error)
	GetNamespace(docID int) (namespace.Namespace, error)
	ListNamespaces() []dbcommon.DocItem
	GetNamespaceArtefacts(id int) ([]artefact.Artefact, error)
	AllNamespaces() []namespace.Namespace

	// nodes
	IndexNode()
	CreateNode(t map[string]interface{}) (int, error)
	InsertNode(n *nodes.Node) (int, error)
	DeleteNode(docID int) error
	UpdateNode(docID int, t map[string]interface{}) error
	GetNode(docID int) (nodes.Node, error)
	GetNodeByKey(key string) (nodes.Node, error)
	ListNodes() []dbcommon.DocItem
	AllNodes() []nodes.Node

	// Organization
	IndexOrganization()
	InsertOrganization(t *organization.Organization) (int, error)
	CreateOrganization(t map[string]interface{}) (int, error)
	DeleteOrganization(docID int) error
	UpdateOrganization(docID int, t map[string]interface{}) error
	GetOrganizationByName(name string) (organization.Organization, error)
	GetOrganizationsByName(name string) ([]organization.Organization, error)
	GetOrganization(docID int) (organization.Organization, error)
	ListOrganizations() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountOrganizations() int
	AllOrganizations() []organization.Organization

	// Pipelines
	IndexPipeline()
	InsertPipeline(t *agenttasks.Pipeline) (int, error)
	CreatePipeline(t map[string]interface{}) (int, error)
	ClonePipeline(t int) (int, error)
	DeletePipeline(docID int) error
	AllUserPipelines(id int) ([]agenttasks.Pipeline, error)
	UpdatePipeline(docID int, t map[string]interface{}) error
	GetPipeline(docID int) (agenttasks.Pipeline, error)
	ListPipelines() []dbcommon.DocItem
	AllPipelines() []agenttasks.Pipeline

	// settings
	IndexSetting()
	InsertSetting(t *setting.Setting) (int, error)
	CreateSetting(t map[string]interface{}) (int, error)
	DeleteSetting(docID int) error
	UpdateSetting(docID int, t map[string]interface{}) error
	GetSettingByKey(name string) (setting.Setting, error)
	GetSettingByUserID(id int) (setting.Setting, error)
	GetSettingsByField(field, name string) ([]setting.Setting, error)
	GetSettingsByKey(name string) ([]setting.Setting, error)
	GetSettingsByUserID(id int) ([]setting.Setting, error)
	GetSetting(docID int) (setting.Setting, error)
	ListSettings() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountSettings() int
	AllSettings() []setting.Setting

	// Storages
	IndexStorage()
	CreateStorage(t map[string]interface{}) (int, error)
	DeleteStorage(docID int) error
	UpdateStorage(docID int, t map[string]interface{}) error
	SearchStorage(name string) (storage.Storage, error)
	GetStorage(docID int) (storage.Storage, error)
	ListStorages() []dbcommon.DocItem
	GetStorageArtefacts(id int) ([]artefact.Artefact, error)
	AllStorages() []storage.Storage

	// Tasks
	IndexTask()
	InsertTask(t *agenttasks.Task) (int, error)
	CreateTask(t map[string]interface{}) (int, error)
	CloneTask(t int) (int, error)
	DeleteTask(docID int) error
	UpdateTask(docID int, t map[string]interface{}) error
	GetTask(docID int) (agenttasks.Task, error)
	GetTaskArtefacts(id int) ([]artefact.Artefact, error)
	ListTasks() []dbcommon.DocItem
	AllTasks() []agenttasks.Task
	AllUserTask(id int) ([]agenttasks.Task, error)

	// Token
	IndexToken()
	InsertToken(t *token.Token) (int, error)
	CreateToken(t map[string]interface{}) (int, error)
	DeleteToken(docID int) error
	UpdateToken(docID int, t map[string]interface{}) error
	GetTokenByKey(name string) (token.Token, error)
	GetTokenByUserID(id int) (token.Token, error)
	GetTokensByField(field, name string) ([]token.Token, error)
	GetTokensByKey(name string) ([]token.Token, error)
	GetTokensByUserID(id int) ([]token.Token, error)
	GetToken(docID int) (token.Token, error)
	ListTokens() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountTokens() int
	AllTokens() []token.Token

	// User
	IndexUser()
	InsertAndSaltUser(t *user.User) (int, error)
	InsertUser(t *user.User) (int, error)
	CreateUser(t map[string]interface{}) (int, error)
	DeleteUser(docID int) error
	UpdateUser(docID int, t map[string]interface{}) error
	SignIn(name, password string) (user.User, error)
	GetUserByName(name string) (user.User, error)
	GetUserByEmail(email string) (user.User, error)
	GetUsersByEmail(email string) ([]user.User, error)
	GetUsersByName(name string) ([]user.User, error)
	// TODO: To replace with a specific collection to index search
	GetUserByIdentity(identity_type, id string) (user.User, error)
	GetUser(docID int) (user.User, error)
	ListUsers() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountUsers() int
	AllUsers() []user.User

	// Plans
	IndexPlan()
	InsertPlan(t *agenttasks.Plan) (int, error)
	CreatePlan(t map[string]interface{}) (int, error)
	ClonePlan(t int) (int, error)
	DeletePlan(docID int) error
	UpdatePlan(docID int, t map[string]interface{}) error
	GetPlan(docID int) (agenttasks.Plan, error)
	ListPlans() []dbcommon.DocItem
	AllPlans() []agenttasks.Plan
}

// For future, now in PoC state will just support
// tiedot
type Database struct {
	Backend string
	DBPath  string
	DBName  string
	Driver  DatabaseDriver
}

var DBInstance *Database

func NewDatabase(backend string) *Database {
	if DBInstance == nil {
		DBInstance = &Database{Backend: backend}
	}
	if backend == "tiedot" {
		fmt.Println("Tiedot backend")
		DBInstance.Driver = tiedot.New(setting.Configuration.DBPath)
	}

	DBInstance.Driver.Init()
	return DBInstance
}

func Instance() *Database {
	return DBInstance
}

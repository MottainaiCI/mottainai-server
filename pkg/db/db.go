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
	"errors"

	"github.com/MottainaiCI/mottainai-server/pkg/artefact"
	arango "github.com/MottainaiCI/mottainai-server/pkg/db/arangodb"
	"github.com/MottainaiCI/mottainai-server/pkg/namespace"
	"github.com/MottainaiCI/mottainai-server/pkg/nodes"
	organization "github.com/MottainaiCI/mottainai-server/pkg/organization"
	"github.com/MottainaiCI/mottainai-server/pkg/secret"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/storage"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	token "github.com/MottainaiCI/mottainai-server/pkg/token"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	webhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
	tiedot "github.com/MottainaiCI/mottainai-server/pkg/db/tiedot"
	anagent "github.com/mudler/anagent"
)

type DatabaseDriver interface {
	Init()
	InsertDoc(string, map[string]interface{}) (string, error)
	FindDoc(string, string) (map[string]interface{}, error)
	DeleteDoc(string, string) error
	UpdateDoc(string, string, map[string]interface{}) error
	ReplaceDoc(string, string, map[string]interface{}) error
	GetDoc(string, string) (map[string]interface{}, error)
	DropColl(string) error
	ListDocs(string) []dbcommon.DocItem
	// Artefacts
	CreateArtefact(map[string]interface{}) (string, error)
	DeleteArtefact(string) error
	UpdateArtefact(string, map[string]interface{}) error
	GetArtefact(string) (artefact.Artefact, error)
	SearchArtefact(string) (artefact.Artefact, error)
	ListArtefacts() []dbcommon.DocItem
	AllArtefacts() []artefact.Artefact

	// Namespaces
	CreateNamespace(t map[string]interface{}) (string, error)
	DeleteNamespace(docID string) error
	UpdateNamespace(docID string, t map[string]interface{}) error
	SearchNamespace(name string) (namespace.Namespace, error)
	GetNamespace(docID string) (namespace.Namespace, error)
	ListNamespaces() []dbcommon.DocItem
	GetNamespaceArtefacts(id string) ([]artefact.Artefact, error)
	AllNamespaces() []namespace.Namespace

	// nodes
	CreateNode(t map[string]interface{}) (string, error)
	InsertNode(n *nodes.Node) (string, error)
	DeleteNode(docID string) error
	UpdateNode(docID string, t map[string]interface{}) error
	GetNode(docID string) (nodes.Node, error)
	GetNodeByKey(key string) (nodes.Node, error)
	ListNodes() []dbcommon.DocItem
	AllNodes() []nodes.Node

	// Organization
	InsertOrganization(t *organization.Organization) (string, error)
	CreateOrganization(t map[string]interface{}) (string, error)
	DeleteOrganization(docID string) error
	UpdateOrganization(docID string, t map[string]interface{}) error
	GetOrganizationByName(name string) (organization.Organization, error)
	GetOrganizationsByName(name string) ([]organization.Organization, error)
	GetOrganization(docID string) (organization.Organization, error)
	ListOrganizations() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountOrganizations() int
	AllOrganizations() []organization.Organization

	// Pipelines
	InsertPipeline(t *agenttasks.Pipeline) (string, error)
	CreatePipeline(t map[string]interface{}) (string, error)
	ClonePipeline(config *setting.Config, t string) (string, error)
	DeletePipeline(docID string) error
	AllUserPipelines(config *setting.Config, id string) ([]agenttasks.Pipeline, error)
	UpdatePipeline(docID string, t map[string]interface{}) error
	GetPipeline(config *setting.Config, docID string) (agenttasks.Pipeline, error)
	ListPipelines() []dbcommon.DocItem
	AllPipelines(config *setting.Config) []agenttasks.Pipeline

	// settings
	InsertSetting(t *setting.Setting) (string, error)
	CreateSetting(t map[string]interface{}) (string, error)
	DeleteSetting(docID string) error
	UpdateSetting(docID string, t map[string]interface{}) error
	GetSettingByKey(name string) (setting.Setting, error)
	GetSettingByUserID(id string) (setting.Setting, error)
	GetSettingsByField(field, name string) ([]setting.Setting, error)
	GetSettingsByKey(name string) ([]setting.Setting, error)
	GetSettingsByUserID(id string) ([]setting.Setting, error)
	GetSetting(docID string) (setting.Setting, error)
	ListSettings() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountSettings() int
	AllSettings() []setting.Setting

	// Storages
	CreateStorage(t map[string]interface{}) (string, error)
	DeleteStorage(docID string) error
	UpdateStorage(docID string, t map[string]interface{}) error
	SearchStorage(name string) (storage.Storage, error)
	GetStorage(docID string) (storage.Storage, error)
	ListStorages() []dbcommon.DocItem
	GetStorageArtefacts(id string) ([]artefact.Artefact, error)
	AllStorages() []storage.Storage

	// Tasks
	InsertTask(t *agenttasks.Task) (string, error)
	CreateTask(t map[string]interface{}) (string, error)
	CloneTask(config *setting.Config, t string) (string, error)
	DeleteTask(config *setting.Config, docID string) error
	UpdateTask(docID string, t map[string]interface{}) error
	GetTask(config *setting.Config, docID string) (agenttasks.Task, error)
	GetTaskArtefacts(id string) ([]artefact.Artefact, error)
	ListTasks() []dbcommon.DocItem
	AllTasks(config *setting.Config) []agenttasks.Task
	AllUserTask(config *setting.Config, id string) ([]agenttasks.Task, error)
	AllNodeTask(config *setting.Config, id string) ([]agenttasks.Task, error)
	GetTaskByStatus(*setting.Config, string) ([]agenttasks.Task, error)

	// Token
	InsertToken(t *token.Token) (string, error)
	CreateToken(t map[string]interface{}) (string, error)
	DeleteToken(docID string) error
	UpdateToken(docID string, t map[string]interface{}) error
	GetTokenByKey(name string) (token.Token, error)
	GetTokenByUserID(id string) (token.Token, error)
	GetTokensByField(field, name string) ([]token.Token, error)
	GetTokensByKey(name string) ([]token.Token, error)
	GetTokensByUserID(id string) ([]token.Token, error)
	GetToken(docID string) (token.Token, error)
	ListTokens() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountTokens() int
	AllTokens() []token.Token

	// User
	InsertAndSaltUser(t *user.User) (string, error)
	InsertUser(t *user.User) (string, error)
	CreateUser(t map[string]interface{}) (string, error)
	DeleteUser(docID string) error
	UpdateUser(docID string, t map[string]interface{}) error
	SignIn(name, password string) (user.User, error)
	GetUserByName(name string) (user.User, error)
	GetUserByEmail(email string) (user.User, error)
	GetUsersByEmail(email string) ([]user.User, error)
	GetUsersByName(name string) ([]user.User, error)
	// TODO: To replace with a specific collection to index search
	GetUserByIdentity(identity_type, id string) (user.User, error)
	GetUser(docID string) (user.User, error)
	ListUsers() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountUsers() int
	AllUsers() []user.User

	// Plans
	InsertPlan(t *agenttasks.Plan) (string, error)
	CreatePlan(t map[string]interface{}) (string, error)
	ClonePlan(config *setting.Config, t string) (string, error)
	DeletePlan(docID string) error
	UpdatePlan(docID string, t map[string]interface{}) error
	GetPlan(config *setting.Config, docID string) (agenttasks.Plan, error)
	ListPlans() []dbcommon.DocItem
	AllPlans(config *setting.Config) []agenttasks.Plan

	// WebHook
	InsertWebHook(t *webhook.WebHook) (string, error)
	CreateWebHook(t map[string]interface{}) (string, error)
	DeleteWebHook(docID string) error
	UpdateWebHook(docID string, t map[string]interface{}) error
	GetWebHookByKey(name string) (webhook.WebHook, error)
	GetWebHookByUserID(id string) (webhook.WebHook, error)
	GetWebHookByURL(id string) (webhook.WebHook, error)

	GetWebHooksByField(field, name string) ([]webhook.WebHook, error)
	GetWebHooksByKey(name string) ([]webhook.WebHook, error)
	GetWebHooksByUserID(id string) ([]webhook.WebHook, error)
	GetWebHooksByURL(id string) ([]webhook.WebHook, error)

	GetWebHook(docID string) (webhook.WebHook, error)
	ListWebHooks() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountWebHooks() int
	AllWebHooks() []webhook.WebHook

	// Secret
	InsertSecret(t *secret.Secret) (string, error)
	CreateSecret(t map[string]interface{}) (string, error)
	DeleteSecret(docID string) error
	UpdateSecret(docID string, t map[string]interface{}) error
	GetSecretByUserID(id string) (secret.Secret, error)
	GetSecretByName(id string) (secret.Secret, error)

	GetSecretsByUserID(id string) ([]secret.Secret, error)
	GetSecretsByName(id string) ([]secret.Secret, error)

	GetSecret(docID string) (secret.Secret, error)
	ListSecrets() []dbcommon.DocItem
	// TODO: Change it, expensive for now
	CountSecrets() int
	AllSecrets() []secret.Secret

	// TODO: See if it's correct expone this as method
	GetAgent() *anagent.Anagent
}

// For future, now in PoC state will just support
// tiedot
type Database struct {
	Backend string
	DBPath  string
	DBName  string
	Driver  DatabaseDriver
	// TODO: Temporary insert Config. See if add a ConfigDatabase object.
	Config *setting.Config
}

var DBInstance *Database

func NewDatabase(config *setting.Config) *Database {
	if DBInstance == nil {
		DBInstance = &Database{Backend: config.GetDatabase().DBEngine, Config: config}
	}
	// TODO: refactor this
	switch config.GetDatabase().DBEngine {
	case "tiedot":
		DBInstance.Driver = tiedot.New(config.GetDatabase().DBPath)
	case "arangodb":
		DBInstance.Driver = arango.New(config.GetDatabase().DatabaseName,
			config.GetDatabase().User, config.GetDatabase().Password,
			config.GetDatabase().CertPath, config.GetDatabase().KeyPath,
			config.GetDatabase().Endpoints)
	default:
		panic(errors.New("Invalid engine defined: '" + config.GetDatabase().DBEngine + "'"))
	}

	if DBInstance.Driver == nil {
		panic(errors.New("Error on initialize Database Driver"))
	}

	DBInstance.Driver.GetAgent().Map(config)
	DBInstance.Driver.Init()
	DBInstance.Config = config
	return DBInstance
}

func Instance() *Database {
	return DBInstance
}

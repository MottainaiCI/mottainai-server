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

package setting

import (
	"fmt"

	v "github.com/spf13/viper"
)

const (
	MOTTAINAI_VERSION    = "0.0000001"
	MOTTAINAI_ENV_PREFIX = "MOTTAINAI"
	MOTTAINAI_CONFIGNAME = "mottainai-server"
	MOTTAINAI_CONFIGPATH = "/etc/mottainai"
)

type Config struct {
	Viper *v.Viper

	// Web UI Settings
	Protocol  string `mapstructure:"webui_protocol"`
	AppSubURL string `mapstructure:"webui_url"`
	HTTPAddr  string `mapstructure:"webui_listenaddress"`
	HTTPPort  string `mapstructure:"webui_port"`
	AppName   string `mapstructure:"application_name"`
	AppURL    string `mapstructure:"application_url"`
	SecretKey string `mapstructure:"secret_key"`

	StaticRootPath string `mapstructure:"root_path"`
	CustomPath     string `mapstructure:"custom_path"`
	DBEngine       string `mapstructure:"db_engine"`
	DBPath         string `mapstructure:"db_path"`
	ArtefactPath   string `mapstructure:"artefact_path"`
	NamespacePath  string `mapstructure:"namespace_path"`
	StoragePath    string `mapstructure:"storage_path"`
	BuildPath      string `mapstructure:"build_path"`
	LockPath       string `mapstructure:"lock_path"`

	ResultsExpireIn int `mapstructure:"results_expire_in"`

	/* Broker Settings */
	Broker                   string            `mapstructure:"broker"`
	BrokerType               string            `mapstructure:"broker_type"`
	BrokerDefaultQueue       string            `mapstructure:"broker_default_queue"`
	BrokerResultBackend      string            `mapstructure:"broker_result_backend"`
	BrokerURI                string            `mapstructure:"broker_uri"`
	BrokerPass               string            `mapstructure:"broker_pass"`
	BrokerUser               string            `mapstructure:"broker_user"`
	BrokerExchange           string            `mapstructure:"broker_exchange"`
	BrokerExchangeType       string            `mapstructure:"broker_exchange_type"`
	BrokerBindingKey         string            `mapstructure:"broker_binding_key"`
	AgentConcurrency         int               `mapstructure:"agent_concurrency"`
	Queues                   map[string]int    `mapstructure:"queues"`
	CacheRegistryCredentials map[string]string `mapstructure:"cache_registry"`

	AgentKey string `mapstructure:"agent_key"`
	ApiKey   string `mapstructure:"api_key"`

	DockerEndpoint    string   `mapstructure:"docker_endpoint"`
	DockerKeepImg     bool     `mapstructure:"docker_keepimg"`
	DockerPriviledged bool     `mapstructure:"docker_privileged"`
	DockerInDocker    bool     `mapstructure:"docker_in_docker"`
	DockerEndpointDiD string   `mapstructure:"docker_in_docker_endpoint"`
	DockerCaps        []string `mapstructure:"docker_caps"`
	DockerCapsDrop    []string `mapstructure:"docker_caps_drop"`
	PrivateQueue      int      `mapstructure:"private_queue"`
	StandAlone        bool     `mapstructure:"standalone"`

	WebHookGitHubToken  string `mapstructure:"github_token"`
	WebHookGitHubSecret string `mapstructure:"github_secret"`

	TLSCert string `mapstructure:"tls_cert"`
	TLSKey  string `mapstructure:"tls_key"`

	AccessControlAllowOrigin string `mapstructure:"access_control_allow_origin"`
}

var (
	Configuration = &Config{Viper: v.New()}
)

func (c *Config) GenDefault() {
	GenDefault(c.Viper)
}

func GenDefault(viper *v.Viper) {

	viper.SetDefault("webui_protocol", "http")
	viper.SetDefault("webui_url", "http://127.0.0.1:9090")
	viper.SetDefault("webui_listenaddress", "127.0.0.1")
	viper.SetDefault("webui_port", "9090")
	viper.SetDefault("application_name", "Mottainai")
	viper.SetDefault("application_url", "http://127.0.0.1:9090")
	viper.SetDefault("secret_key", "vvH5oXJCTwHNGcMe2EJWDUKg9yY6qx")

	viper.SetDefault("root_path", "./")
	viper.SetDefault("custom_path", "./")
	viper.SetDefault("db_engine", "tiedot")
	viper.SetDefault("db_path", "./.DB")
	viper.SetDefault("artefact_path", "./artefact")
	viper.SetDefault("namespace_path", "./namespace")
	viper.SetDefault("storage_path", "./storage")
	viper.SetDefault("build_path", "/build/")
	viper.SetDefault("lock_path", "/var/lock/mottainai/")

	viper.SetDefault("results_expire_in", 3600)

	viper.SetDefault("broker", "amqp://guest@127.0.0.1:5672/")
	viper.SetDefault("broker_type", "amqp")
	viper.SetDefault("broker_default_queue", "global_tasks")
	viper.SetDefault("broker_result_backend", "amqp://guest@127.0.0.1:5672/")
	viper.SetDefault("broker_uri", "http://127.0.0.1:15672")
	viper.SetDefault("broker_pass", "guest")
	viper.SetDefault("broker_user", "guest")
	viper.SetDefault("broker_exchange", "machinery_exchange")
	viper.SetDefault("broker_exchange_type", "direct")
	viper.SetDefault("broker_binding_key", "machinery_task")
	viper.SetDefault("agent_concurrency", 1)
	viper.SetDefault("queues", map[string]int{})
	viper.SetDefault("cache_registry", map[string]int{})

	viper.SetDefault("agent_key", "")
	viper.SetDefault("api_key", "")

	viper.SetDefault("docker_endpoint", "unix:///var/run/docker.sock")
	viper.SetDefault("docker_keepimg", true)
	viper.SetDefault("docker_privileged", false)
	viper.SetDefault("docker_in_docker", false)
	viper.SetDefault("docker_in_docker_endpoint", "/var/run/docker.sock")
	viper.SetDefault("docker_caps", []string{"SYS_PTRACE"})
	viper.SetDefault("docker_caps_drop", []string{})
	viper.SetDefault("private_queue", 1)
	viper.SetDefault("standalone", false)
	viper.SetDefault("github_token", "")
	viper.SetDefault("github_secret", "")

	viper.SetDefault("tls_cert", "")
	viper.SetDefault("tls_key", "")

	viper.SetDefault("access_control_allow_origin", "*")
}

func (c *Config) Unmarshal() error {
	var err error

	if Configuration.Viper.InConfig("etcd-config") &&
		Configuration.Viper.GetBool("etcd-config") {
		err = Configuration.Viper.ReadRemoteConfig()
	} else {
		err = Configuration.Viper.ReadInConfig()
		// TODO: add loglevel warning related to no config file processed
	}

	err = Configuration.Viper.Unmarshal(&Configuration)

	return err
}

func (c *Config) String() string {
	// TODO: Currently I don't find a way to create a json from
	//       with viper to a io.Writer (or string)
	var ans string = fmt.Sprintf(`
configfile: %s

webui_protocol: %s
webui_url: %s
webui_listenaddress: %s
webui_port: %s
application_name: %s
application_url: %s
secret_key: **************

root_path: %s
custom_path: %s
db_engine: %s
db_path: %s
artefact_path: %s
namespace_path: %s
storage_path: %s
build_path: %s
lock_path: %s

results_expire_in: %d

broker: %s
broker_type: %s
broker_default_queue: %s
broker_result_backend: %s
broker_uri: %s
broker_pass: %s
broker_user: %s
broker_exchange: %s
broker_exchange_type: %s
broker_binding_key: %s
agent_concurrency: %d
queues: %v
cache_registry: %v
agent_key: ***************
api_key: ***************

docker_endpoint: %s
docker_keepimg: %t
docker_privileged: %t
docker_in_docker: %t
docker_in_docker_endpoint: %s
docker_caps: %s
docker_caps_drop: %s
private_queue: %d
standalone: %t
github_token: %s
github_secret: *****************

tls_cert: %s
tls_key: ***********************

access_control_allow_origin: %s
`,
		c.Viper.Get("config"),
		c.Protocol, c.AppSubURL, c.HTTPAddr, c.HTTPPort, c.AppName, c.AppURL,
		c.StaticRootPath, c.CustomPath, c.DBEngine, c.DBPath, c.ArtefactPath,
		c.NamespacePath, c.StoragePath, c.BuildPath, c.LockPath,
		c.ResultsExpireIn,
		c.Broker, c.BrokerType, c.BrokerDefaultQueue, c.BrokerResultBackend,
		c.BrokerURI, c.BrokerPass, c.BrokerUser, c.BrokerExchange, c.BrokerExchangeType,
		c.BrokerBindingKey, c.AgentConcurrency, c.Queues, c.CacheRegistryCredentials,
		c.DockerEndpoint, c.DockerKeepImg, c.DockerPriviledged, c.DockerInDocker,
		c.DockerEndpointDiD, c.DockerCaps, c.DockerCapsDrop, c.PrivateQueue, c.StandAlone,
		c.WebHookGitHubToken, c.TLSCert, c.AccessControlAllowOrigin)

	return ans
}

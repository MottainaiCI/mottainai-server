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

const MOTTAINAI_VERSION = "0.0000001"

type Config struct {
	Protocol  string `yaml:"webui_protocol" envconfig:"WEBUI_PROTOCOL"`
	AppSubURL string `yaml:"webui_url" envconfig:"WEBUI_URL"`
	HTTPAddr  string `yaml:"webui_listenaddress" envconfig:"WEBUI_LISTENADDRESS"`
	HTTPPort  string `yaml:"webui_port" envconfig:"WEBUI_PORT"`
	AppName   string `yaml:"application_name" envconfig:"APPLICATION_NAME"`
	AppURL    string `yaml:"application_url" envconfig:"APPLICATION_URL"`
	SecretKey string `yaml:"secret_key" envconfig:"SECRET_KEY"`

	StaticRootPath string `yaml:"root_path" envconfig:"ROOT_PATH"`
	CustomPath     string `yaml:"custom_path" envconfig:"CUSTOM_PATH"`
	DBEngine       string `yaml:"db_engine" envconfig:"DB_ENGINE"`
	DBPath         string `yaml:"db_path" envconfig:"DB_PATH"`
	ArtefactPath   string `yaml:"artefact_path" envconfig:"ARTEFACT_PATH"`
	NamespacePath  string `yaml:"namespace_path" envconfig:"NAMESPACE_PATH"`
	StoragePath    string `yaml:"storage_path" envconfig:"STORAGE_PATH"`
	BuildPath      string `yaml:"build_path" envconfig:"BUILD_PATH"`

	ResultsExpireIn int `yaml:"results_expire_in" envconfig:"RESULTS_EXPIRE_IN"`

	/* AMQP Settings */

	AMQPBroker        string `yaml:"amqp_broker" envconfig:"AMQP_BROKER"`
	AMQPDefaultQueue  string `yaml:"amqp_default_queue" envconfig:"AMQP_DEFAULT_QUEUE"`
	AMQPResultBackend string `yaml:"amqp_result_backend" envconfig:"AMQP_RESULT_BACKEND"`
	AMQPURI           string `yaml:"amqp_uri" envconfig:"AMQP_URI"`
	AMQPPass          string `yaml:"amqp_pass" envconfig:"AMQP_PASS"`
	AMQPUser          string `yaml:"amqp_user" envconfig:"AMQP_USER"`
	AMQPExchange      string `yaml:"amqp_exchange" envconfig:"AMQP_EXCHANGE"`
	AMQPExchangeType  string `yaml:"amqp_exchange_type" envconfig:"AMQP_EXCHANGE_TYPE"`
	AMQPBindingKey    string `yaml:"amqp_binding_key" envconfig:"AMQP_BINDING_KEY"`
	AgentConcurrency  int    `yaml:"agent_concurrency" envconfig:"AGENT_CONCURRENCY"`

	AgentKey string `yaml:"agent_key" envconfig:"AGENT_KEY"`

	TempWorkDir string `yaml:"work_dir" envconfig:"WORKING_DIR"`

	DockerEndpoint    string `yaml:"docker_endpoint" envconfig:"DOCKER_ENDPOINT"`
	DockerKeepImg     bool   `yaml:"docker_keepimg" envconfig:"DOCKER_KEEPIMG"`
	DockerPriviledged bool   `yaml:"docker_privileged" envconfig:"DOCKER_PRIVILEGED"`
	DockerInDocker    bool   `yaml:"docker_in_docker" envconfig:"DOCKER_IN_DOCKER"`
	DockerEndpointDiD string `yaml:"docker_in_docker_endpoint" envconfig:"DOCKER_IN_DOCKER_ENDPOINT"`
}

var (
	AppVer string

	Configuration = &Config{}

	Protocol                   string
	AppSubURL                  string
	HTTPAddr                   string
	HTTPPort                   string
	AppName                    string
	AppURL                     string
	SecretKey                  string
	TimeFormat                 string
	ShowFooterTemplateLoadTime bool
	UI                         string
	StaticRootPath             string
	ArtefactPath               string
	NamespacePath              string
	StoragePath                string
	BuildPath                  string

	CustomPath string
	DBEngine   string
	DBPath     string

	/* AMQP Settings */

	AMQPBroker        string
	AMQPDefaultQueue  string
	AMQPResultBackend string
	AMQPURI           string
	AMQPPass          string
	AMQPUser          string
	AMQPExchange      string
	AMQPExchangeType  string
	AMQPBindingKey    string
	AgentConcurrency  int
	ResultsExpireIn   int
	AgentKey          string

	TempWorkDir string

	DockerEndpoint    string
	DockerKeepImg     bool
	DockerPriviledged bool
	DockerInDocker    bool
	DockerEndpointDiD string
)

func GenDefault() {

	AppVer = MOTTAINAI_VERSION
	Configuration.HTTPAddr = "127.0.0.1"
	Configuration.HTTPPort = "9090"
	Configuration.Protocol = "http"
	Configuration.AppName = "Mottainai"
	Configuration.AppURL = "http://127.0.0.1:9090"
	Configuration.SecretKey = "vvH5oXJCTwHNGcMe2EJWDUKg9yY6qx"
	Configuration.StaticRootPath = "./"
	Configuration.ArtefactPath = "./artefact"
	Configuration.NamespacePath = "./namespace"
	Configuration.StoragePath = "./storage"
	Configuration.BuildPath = "/build/"

	Configuration.CustomPath = "./"
	Configuration.AppSubURL = "http://127.0.0.1:9090/"
	Configuration.DBEngine = "tiedot"
	Configuration.DBPath = "./.DB"

	Configuration.AMQPBroker = "amqp://guest@127.0.0.1:5672/"
	Configuration.AMQPDefaultQueue = "global_tasks"
	Configuration.AMQPExchange = "machinery_exchange"

	Configuration.AMQPURI = "http://127.0.0.1:15672"
	Configuration.AMQPUser = "guest"
	Configuration.AMQPPass = "guest"
	Configuration.ResultsExpireIn = 3600
	Configuration.AMQPResultBackend = "amqp://guest@127.0.0.1:5672/"
	Configuration.AMQPExchange = "machinery_exchange"
	Configuration.AMQPExchangeType = "direct"
	Configuration.AMQPBindingKey = "machinery_task"
	Configuration.AgentConcurrency = 1

	// TODO: to remove from here, needs to be setted only on the agent side.
	Configuration.AgentKey = ""

	Configuration.TempWorkDir = "/tmp"
	Configuration.DockerEndpoint = "unix:///var/run/docker.sock"
	Configuration.DockerKeepImg = true
	Configuration.DockerPriviledged = true
	Configuration.DockerInDocker = true
	Configuration.DockerEndpointDiD = "/var/run/docker.sock"

	LoadFromEnvironment()
}

func LoadFromFileEnvironment(cnfPath string) error {
	cfg, err := fromFile(cnfPath)
	if err == nil {
		Configuration = cfg
	} else {
		return err
	}

	return LoadFromEnvironment()
}

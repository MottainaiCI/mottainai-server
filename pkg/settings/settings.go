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
	"errors"
	"fmt"
	"strings"

	v "github.com/spf13/viper"
)

const (
	MOTTAINAI_VERSION    = "0.1"
	MOTTAINAI_ENV_PREFIX = "MOTTAINAI"
	MOTTAINAI_CONFIGNAME = "mottainai-server"
	MOTTAINAI_CONFIGPATH = "/etc/mottainai"

	Timeformat = "20060102150405"
)

// Web UI Settings
type WebConfig struct {
	Protocol  string `mapstructure:"protocol"`
	AppSubURL string `mapstructure:"url"`
	HTTPAddr  string `mapstructure:"listenaddress"`
	HTTPPort  string `mapstructure:"port"`

	AppName              string `mapstructure:"application_name"`
	AppBrandingLogo      string `mapstructure:"application_branding_logo"`
	AppBrandingLogoSmall string `mapstructure:"application_branding_logo_small"`
	AppBrandingFavicon   string `mapstructure:"application_branding_favicon"`

	// TODO: TO rename in API URL
	AppURL string `mapstructure:"application_url"`

	// Replate old custom_path
	TemplatePath string `mapstructure:"template_path"`

	StaticRootPath string `mapstructure:"root_path"`

	AccessControlAllowOrigin string `mapstructure:"access_control_allow_origin"`

	// WebHook Parameters
	EmbedWebHookServer     bool   `mapstructure:"embed_webhookserver"`
	AccessToken            string `mapstructure:"access_token"`
	WebHookGitHubToken     string `mapstructure:"github_token"`
	WebHookGitHubTokenUser string `mapstructure:"github_token_user"`
	WebHookGitHubSecret    string `mapstructure:"github_secret"`
	WebHookToken           string `mapstructure:"webhook_token"`

	LockPath     string `mapstructure:"lock_path"`
	UploadTmpDir string `mapstructure:"upload_tmpdir"`

	HealthCheckInterval int `mapstructure:"healthcheck_interval"`
	TaskDeadline        int `mapstructure:"task_deadline"`
	NodeDeadline        int `mapstructure:"node_deadline"`

	SessionProvider       string `mapstructure:"session_provider"`
	SessionProviderConfig string `mapstructure:"session_provider_config"`
}

type StorageConfig struct {
	Type string `mapstructure:"type"`

	ArtefactPath  string `mapstructure:"artefact_path"`
	NamespacePath string `mapstructure:"namespace_path"`
	StoragePath   string `mapstructure:"storage_path"`
}

type DatabaseConfig struct {
	DBEngine string `mapstructure:"engine"`
	DBPath   string `mapstructure:"db_path"`

	Endpoints    []string `mapstructure:"db_endpoints"`
	User         string   `mapstructure:"db_user"`
	DatabaseName string   `mapstructure:"db_name"`
	Password     string   `mapstructure:"db_password"`
	CertPath     string   `mapstructure:"db_certpath"`
	KeyPath      string   `mapstructure:"db_keypath"`
}

type BrokerConfig struct {
	Type string `mapstructure:"type"`

	ResultsExpireIn int `mapstructure:"results_expire_in"`

	HandleSignal bool `mapstructure:"handle_signal"`

	/* Broker Settings */
	Broker              string `mapstructure:"broker"`
	BrokerDefaultQueue  string `mapstructure:"default_queue"`
	BrokerResultBackend string `mapstructure:"result_backend"`
	BrokerURI           string `mapstructure:"mgmt_uri"`
	BrokerPass          string `mapstructure:"pass"`
	BrokerUser          string `mapstructure:"user"`
	BrokerExchange      string `mapstructure:"exchange"`
	BrokerExchangeType  string `mapstructure:"exchange_type"`
	BrokerBindingKey    string `mapstructure:"binding_key"`

	// Redis
	MaxIdle                int  `mapstructure:"max_idle"`
	MaxActive              int  `mapstructure:"max_active"`
	IdleTimeout            int  `mapstructure:"idle_timeout"`
	Wait                   bool `mapstructure:"wait"`
	ReadTimeout            int  `mapstructure:"read_timeout"`
	WriteTimeout           int  `mapstructure:"write_timeout"`
	ConnectTimeout         int  `mapstructure:"connect_timeout"`
	DelayedTasksPollPeriod int  `mapstructure:"delayed_tasks_poll_period"`

	// DynamoDB
	TaskStatesTable string `mapstructure:"task_states_table"`
	GroupMetasTable string `mapstructure:"group_metas_table"`
}

type AgentConfig struct {
	SecretKey          string         `mapstructure:"secret_key"`
	BuildPath          string         `mapstructure:"build_path"`
	AgentConcurrency   int            `mapstructure:"concurrency"`
	AgentKey           string         `mapstructure:"agent_key"`
	ApiKey             string         `mapstructure:"api_key"`
	PrivateQueue       int            `mapstructure:"private_queue"`
	StandAlone         bool           `mapstructure:"standalone"`
	DownloadRateLimit  int64          `mapstructure:"download_speed_limit"`
	UploadRateLimit    int64          `mapstructure:"upload_speed_limit"`
	Queues             map[string]int `mapstructure:"queues"`
	UploadChunkSize    int            `mapstructure:"upload_chunk_size"`
	SupportedExecutors []string       `mapstructure:"executor"`

	// List of command to execute before execute a task
	PreTaskHookExec []string `mapstructure:"pre_task_hook_exec"`

	DockerEndpoint    string   `mapstructure:"docker_endpoint"`
	DockerKeepImg     bool     `mapstructure:"docker_keepimg"`
	DockerPriviledged bool     `mapstructure:"docker_privileged"`
	DockerInDocker    bool     `mapstructure:"docker_in_docker"`
	DockerEndpointDiD string   `mapstructure:"docker_in_docker_endpoint"`
	DockerCaps        []string `mapstructure:"docker_caps"`
	DockerCapsDrop    []string `mapstructure:"docker_caps_drop"`
	DefaultTaskQuota  string   `mapstructure:"default_task_quota"`

	KubeConfigPath   string `mapstructure:"kubeconfig"`
	KubeNamespace    string `mapstructure:"kube_namespace"`
	KubeStorageClass string `mapstructure:"kube_storageclass"`
	KubeDropletImage string `mapstructure:"kube_droplet_image"`

	LxdEndpoint            string            `mapstructure:"lxd_endpoint"`
	LxdConfigDir           string            `mapstructure:"lxd_config_dir"`
	LxdProfiles            []string          `mapstructure:"lxd_profiles"`
	LxdEphemeralContainers bool              `mapstructure:"lxd_ephemeral_containers"`
	LxdCacheRegistry       map[string]string `mapstructure:"lxd_cache_registry"`

	CacheRegistryCredentials map[string]string `mapstructure:"cache_registry"`

	HealthCheckExec      []string `mapstructure:"health_check_exec"`
	HealthCheckCleanPath []string `mapstructure:"health_check_clean_path"`
}

type GeneralConfig struct {
	Debug         bool   `mapstructure:"debug"`
	LogFile       string `mapstructure:"logfile"`
	LogLevel      string `mapstructure:"loglevel"`
	TLSCert       string `mapstructure:"tls_cert"`
	TLSKey        string `mapstructure:"tls_key"`
	ClientTimeout int    `mapstructure:"client_timeout"`
}

type Config struct {
	Viper *v.Viper

	General  GeneralConfig  `mapstructure:"general"`
	Web      WebConfig      `mapstructure:"web"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Database DatabaseConfig `mapstructure:"db"`
	Broker   BrokerConfig   `mapstructure:"broker"`
	Agent    AgentConfig    `mapstructure:"agent"`
}

func (c *Config) GetWeb() *WebConfig {
	return &c.Web
}

func (c *Config) GetStorage() *StorageConfig {
	return &c.Storage
}

func (c *Config) GetDatabase() *DatabaseConfig {
	return &c.Database
}

func (c *Config) GetBroker() *BrokerConfig {
	return &c.Broker
}

func (c *Config) GetAgent() *AgentConfig {
	return &c.Agent
}

func (c *Config) GetGeneral() *GeneralConfig {
	return &c.General
}

func (c *Config) GenDefault() {
	GenDefault(c.Viper)
}

func NewConfig(viper *v.Viper) *Config {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &Config{Viper: viper}
}

func GenDefault(viper *v.Viper) {

	viper.SetDefault("web.protocol", "http")
	viper.SetDefault("web.url", "/")
	viper.SetDefault("web.listenaddress", "127.0.0.1")
	viper.SetDefault("web.port", "9090")
	viper.SetDefault("web.application_name", "Mottainai")
	viper.SetDefault("web.application_url", "http://127.0.0.1:9090")
	viper.SetDefault("web.template_path", "./")
	viper.SetDefault("web.root_path", "./")
	viper.SetDefault("web.access_control_allow_origin", "*")
	viper.SetDefault("web.embed_webhookserver", true)
	viper.SetDefault("web.access_token", "")
	viper.SetDefault("web.github_token", "")
	viper.SetDefault("web.github_secret", "")
	viper.SetDefault("web.github_token_user", "")
	viper.SetDefault("web.webhook_token", "")
	viper.SetDefault("web.lock_path", "/srv/mottainai/lock")
	viper.SetDefault("web.upload_tmpdir", "/var/tmp")
	viper.SetDefault("web.task_deadline", 21600) // 6h
	viper.SetDefault("web.node_deadline", 21600)
	viper.SetDefault("web.healthcheck_interval", 800)
	viper.SetDefault("web.session_provider", "")
	viper.SetDefault("web.session_provider_config", "")

	viper.SetDefault("storage.type", "dir")
	viper.SetDefault("storage.artefact_path", "./artefact")
	viper.SetDefault("storage.namespace_path", "./namespace")
	viper.SetDefault("storage.storage_path", "./storage")

	viper.SetDefault("db.engine", "tiedot")
	viper.SetDefault("db.db_path", "./.DB")
	viper.SetDefault("db.db_endpoints", []string{})
	viper.SetDefault("db.db_user", "")
	viper.SetDefault("db.db_name", "")
	viper.SetDefault("db.db_password", "")
	viper.SetDefault("db.db_certpath", "")
	viper.SetDefault("db.db_keypath", "")

	viper.SetDefault("broker.handle_signal", true)
	viper.SetDefault("broker.type", "amqp")
	viper.SetDefault("broker.results_expire_in", 3600)
	viper.SetDefault("broker.broker", "amqp://guest@127.0.0.1:5672/")
	viper.SetDefault("broker.default_queue", "global_tasks")
	viper.SetDefault("broker.result_backend", "amqp://guest@127.0.0.1:5672/")
	viper.SetDefault("broker.mgmt_uri", "")
	viper.SetDefault("broker.pass", "")
	viper.SetDefault("broker.user", "")
	viper.SetDefault("broker.exchange", "machinery_exchange")
	viper.SetDefault("broker.exchange_type", "direct")
	viper.SetDefault("broker.binding_key", "machinery_task")

	viper.SetDefault("agent.secret_key", "vvH5oXJCTwHNGcMe2EJWDUKg9yY6qx")
	viper.SetDefault("agent.build_path", "/srv/mottainai/build")
	viper.SetDefault("agent.concurrency", 1)
	viper.SetDefault("agent.agent_key", "")
	viper.SetDefault("agent.api_key", "")
	viper.SetDefault("agent.private_queue", 1)
	viper.SetDefault("agent.standalone", false)
	viper.SetDefault("agent.upload_speed_limit", 0)
	viper.SetDefault("agent.download_speed_limit", 0)
	viper.SetDefault("agent.upload_chunk_size", 512)

	viper.SetDefault("agent.queues", map[string]int{})
	viper.SetDefault("agent.cache_registry", map[string]int{})

	viper.SetDefault("agent.docker_endpoint", "unix:///var/run/docker.sock")
	viper.SetDefault("agent.docker_keepimg", true)
	viper.SetDefault("agent.docker_privileged", false)
	viper.SetDefault("agent.docker_in_docker", false)
	viper.SetDefault("agent.docker_in_docker_endpoint", "/var/run/docker.sock")
	viper.SetDefault("agent.docker_caps", []string{"SYS_PTRACE"})
	viper.SetDefault("agent.docker_caps_drop", []string{})
	viper.SetDefault("agent.kubeconfig", "")
	viper.SetDefault("agent.kube_namespace", "default")
	viper.SetDefault("agent.kube_storageclass", "standard")
	viper.SetDefault("agent.default_task_quota", "100Gi")
	viper.SetDefault("agent.kube_droplet_image", "busybox:latest")

	viper.SetDefault("agent.lxd_endpoint", "")
	viper.SetDefault("agent.lxd_config_dir", "/srv/mottainai/lxc/")
	viper.SetDefault("agent.lxd_ephemeral_containers", true)
	viper.SetDefault("agent.lxd_profiles", []string{})
	viper.SetDefault("agent.lxd_cache_registry", map[string]int{})

	viper.SetDefault("agent.health_check_clean_path", []string{})
	viper.SetDefault("agent.health_check_exec", []string{})

	viper.SetDefault("agent.pre_task_hook_exec", []string{})
	viper.SetDefault("agent.executor", []string{})

	viper.SetDefault("general.tls_cert", "")
	viper.SetDefault("general.tls_key", "")
	viper.SetDefault("general.debug", false)
	viper.SetDefault("general.logfile", "")
	viper.SetDefault("general.client_timeout", 360)
	viper.SetDefault("general.loglevel", "info")
}

func (c *Config) Unmarshal() error {
	var err error

	if c.Viper.InConfig("etcd-config") &&
		c.Viper.GetBool("etcd-config") {
		err = c.Viper.ReadRemoteConfig()
	} else {
		err = c.Viper.ReadInConfig()
		// TODO: add loglevel warning related to no config file processed
	}

	err = c.Viper.Unmarshal(&c)

	return err
}

func (c *WebConfig) GetProtocol() string {
	return c.Protocol
}

func (c *WebConfig) BuildAbsURL(pattern string) string {
	path := strings.TrimRight(c.AppURL, "/")
	return path + pattern
}

func (c *WebConfig) BuildURI(pattern string) string {
	var path string = c.AppSubURL
	if path[len(path)-1:] == "/" {
		if len(path) == 1 {
			path = ""
		} else {
			path = path[0 : len(path)-1]
		}
	}
	if len(pattern) != 0 && pattern[0:1] != "/" {
		pattern = "/" + pattern
	}
	return path + pattern
}

func (c *WebConfig) CompareURI(requestURI, pattern string) bool {
	// TODO: Complete handle of complete URL with schema http://...

	url := c.BuildURI(pattern)
	if url == requestURI {
		return true
	}
	return false
}

func (c *WebConfig) HasPrefixURL(requestURI, prefix string) bool {
	// TODO: Complete handle of complete URL with schema http://...

	url := c.BuildURI(prefix)

	if strings.HasPrefix(requestURI, url) {
		return true
	}

	return false
}

/*
   Return path of resource without application prefix.
*/
func (c *WebConfig) NormalizePath(requestPath string) (string, error) {
	var ans = requestPath
	if c.AppSubURL == "/" {
		return requestPath, nil
	}

	if len(c.AppSubURL) > len(requestPath) {
		return "", errors.New("Invalid path")
	}

	if strings.HasPrefix(requestPath, c.AppSubURL) {
		ans = requestPath[len(c.AppSubURL):]
		if !strings.HasPrefix(ans, "/") {
			ans = "/" + ans
		}
	}

	return ans, nil
}

func (c *WebConfig) GroupAppPath() string {
	var ans string
	if c.AppSubURL == "/" {
		ans = ""
	} else if strings.HasSuffix(c.AppSubURL, "/") {
		ans = c.AppSubURL[0 : len(c.AppSubURL)-1]
	} else {
		ans = c.AppSubURL
	}

	return ans
}

func (c *WebConfig) String() string {
	var ans string = fmt.Sprintf(`
web:
  protocol: %s
  url: %s
  listenaddress: %s
  port: %s
  application_name: %s
  application_branding_logo: %s
  application_branding_logo_small: %s
  application_branding_favicon: %s

  application_url: %s

  template_path: %s

  access_control_allow_origin: %s

  embed_webhookserver: %v
  access_token: %s
  github_token: %s
  github_token_user: %s
  github_secret: %s
  webhook_token: %s

  lock_path: %s

  task_deadline: %d
  node_deadline: %d
  healthcheck_interval: %d
`,
		c.Protocol, c.AppSubURL,
		c.HTTPAddr, c.HTTPPort,
		c.AppName, c.AppBrandingLogo, c.AppBrandingLogoSmall, c.AppBrandingFavicon,
		c.AppURL,
		c.TemplatePath,
		c.AccessControlAllowOrigin,
		c.EmbedWebHookServer,
		c.AccessToken,
		c.WebHookGitHubToken,
		c.WebHookGitHubTokenUser,
		c.WebHookGitHubSecret,
		c.WebHookToken,
		c.LockPath,
		c.TaskDeadline, c.NodeDeadline, c.HealthCheckInterval)

	return ans
}

func (c *StorageConfig) String() string {
	var ans string = fmt.Sprintf(`
storage:
  type: %s
  artefact_path: %s
  namespace_path: %s
  storage_path: %s
`,
		c.Type, c.ArtefactPath,
		c.NamespacePath, c.StoragePath)

	return ans
}

func (c *DatabaseConfig) String() string {
	var ans string = fmt.Sprintf(`
db:
  engine: %s
  db_path: %s
  db_endpoints: %s
  db_name: %s
  db_password: ****
  db_certpath: %s
  db_keypath: %s
  db_user: %s
`,
		c.DBEngine, c.DBPath, c.Endpoints, c.DatabaseName, c.CertPath, c.KeyPath, c.User)
	return ans
}

func (c *BrokerConfig) String() string {
	var ans string = fmt.Sprintf(`
broker:
  handle_signal: %v
  type: %s
  results_expire_in: %d
  broker: %s
  default_queue: %s
  result_backend: %s
  mgmt_uri: %s
  pass: %s
  user: %s
  exchange: %s
  exchange_type: %s
  binding_key: %s

  // Redis only
  max_idle: %d
  max_active: %d
  idle_timeout: %d
  wait: %v
  read_timeout: %d
  write_timeout: %d
  connect_timeout: %d
  delayed_tasks_poll_period: %d

  // DynamoDB only
  task_states_table: %s
  group_metas_table: %s
`,
		c.HandleSignal, c.Type, c.ResultsExpireIn, c.Broker,
		c.BrokerDefaultQueue, c.BrokerResultBackend,
		c.BrokerURI, c.BrokerPass,
		c.BrokerUser, c.BrokerExchange,
		c.BrokerExchangeType, c.BrokerBindingKey,
		c.MaxIdle, c.MaxActive, c.IdleTimeout,
		c.Wait, c.ReadTimeout, c.WriteTimeout,
		c.ConnectTimeout, c.DelayedTasksPollPeriod, c.TaskStatesTable, c.GroupMetasTable)

	return ans
}

func (c *AgentConfig) String() string {
	var ans string = fmt.Sprintf(`
agent:
  secret_key: %s
  build_path: %s
  concurrency: %d
  agent_key: %s
  api_key: %s
  private_queue: %d
  standalone: %t
  download_speed_limit: %d
  upload_speed_limit: %d
  queues: %v
  upload_chunk_size: %d

  docker_endpoint: %s
  docker_keepimg: %t
  docker_privileged: %t
  docker_in_docker: %t
  docker_in_docker_endpoint: %s
  docker_caps: %s
  docker_caps_drop: %s

  lxd_endpoint: %s
  lxd_config_dir: %s
  lxd_profiles: %s
  lxd_ephemeral_containers: %t
  lxd_cache_registry: %s

  cache_registry: %s
  health_check_exec: %s
  health_check_clean_path: %s

  pre_task_hook_exec: %s
`, c.SecretKey, c.BuildPath,
		c.AgentConcurrency, c.AgentKey, c.ApiKey,
		c.PrivateQueue, c.StandAlone, c.DownloadRateLimit,
		c.UploadRateLimit, c.Queues, c.UploadChunkSize,
		c.DockerEndpoint, c.DockerKeepImg,
		c.DockerPriviledged, c.DockerInDocker,
		c.DockerEndpointDiD, c.DockerCaps, c.DockerCapsDrop,
		c.LxdEndpoint, c.LxdConfigDir, c.LxdProfiles, c.LxdEphemeralContainers,
		c.LxdCacheRegistry, c.CacheRegistryCredentials,
		c.HealthCheckExec, c.HealthCheckCleanPath,
		c.PreTaskHookExec)

	return ans
}

func (c *GeneralConfig) String() string {
	var ans string = fmt.Sprintf(`
general:
  debug: %t
  logfile: %s
  loglevel: %s
  tls_cert: %s
  tls_key: ***********************
  client_timeout: %d
`,
		c.Debug, c.LogFile, c.LogLevel,
		c.TLSCert, c.ClientTimeout)

	return ans
}

func (c *Config) String() string {
	// TODO: Currently I don't find a way to create a json from
	//       with viper to a io.Writer (or string)
	var ans string = fmt.Sprintf(`
configfile: %s

%s

%s

%s

%s

%s

%s
`,
		c.Viper.Get("config"),
		c.Web.String(),
		c.Broker.String(),
		c.Storage.String(),
		c.Agent.String(),
		c.Database.String(),
		c.General.String())

	return ans
}

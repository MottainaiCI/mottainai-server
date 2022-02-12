/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
              2021 Daniele Rondina <geaaru@funtoo.org>
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
	"strings"

	v "github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	MOTTAINAI_VERSION    = "0.1.3"
	MOTTAINAI_ENV_PREFIX = "MOTTAINAI"
	MOTTAINAI_CONFIGNAME = "mottainai-server"
	MOTTAINAI_CONFIGPATH = "/etc/mottainai"

	Timeformat = "20060102150405"
)

var (
	BuildTime   string
	BuildCommit string
)

// Web UI Settings
type WebConfig struct {
	Protocol  string `mapstructure:"protocol" json:"protocol,omitempty" yaml:"protocol,omitempty"`
	AppSubURL string `mapstructure:"url" json:"url,omitempty" yaml:"url,omitempty"`
	HTTPAddr  string `mapstructure:"listenaddress" json:"listenaddress,omitempty" yaml:"listenaddress,omitempty"`
	HTTPPort  string `mapstructure:"port" json:"port,omitempty" yaml:"port,omitempty"`

	AppName              string `mapstructure:"application_name" json:"application_name,omitempty" yaml:"application_name,omitempty"`
	AppBrandingLogo      string `mapstructure:"application_branding_logo" json:"application_branding_logo,omitempty" yaml:"application_branding_logo,omitempty"`
	AppBrandingLogoSmall string `mapstructure:"application_branding_logo_small" json:"application_branding_logo_small,omitempty" yaml:"application_branding_logo_small,omitempty"`
	AppBrandingFavicon   string `mapstructure:"application_branding_favicon" json:"application_branding_favicon,omitempty" yaml:"application_branding_favicon,omitempty"`

	// TODO: TO rename in API URL
	AppURL string `mapstructure:"application_url" json:"application_url,omitempty" yaml:"application_url,omitempty"`

	// Replate old custom_path
	TemplatePath string `mapstructure:"template_path" json:"template_path,omitempty" yaml:"template_path,omitempty"`

	StaticRootPath string `mapstructure:"root_path" json:"root_path,omitempty" yaml:"root_path,omitempty"`

	AccessControlAllowOrigin string `mapstructure:"access_control_allow_origin" json:"access_control_allow_origin,omitempty" yaml:"access_control_allow_origin,omitempty"`

	// WebHook Parameters
	EmbedWebHookServer     bool   `mapstructure:"embed_webhookserver" json:"embed_webhookserver,omitempty" yaml;"embed_webhookserver,omitempty"`
	AccessToken            string `mapstructure:"access_token" json:"access_token,omitempty" yaml:"access_token,omitempty"`
	WebHookGitHubToken     string `mapstructure:"github_token" json:"github_token,omitempty" yaml:"github_token,omitempty"`
	WebHookGitHubTokenUser string `mapstructure:"github_token_user" json:"github_token_user,omitempty" yaml:"github_token_user,omitempty"`
	WebHookGitHubSecret    string `mapstructure:"github_secret" json:"github_secret,omitempty" yaml:"github_secret,omitempty"`
	WebHookGitHubCallback  string `mapstructure:"github_callback" json:"github_callback,omitempty" yaml:"github_callback,omitempty"`
	WebHookToken           string `mapstructure:"webhook_token" json:"webhook_token,omitempty" yaml:"webhook_token,omitempty"`

	LockPath     string `mapstructure:"lock_path" json:"lock_path,omitempty" yaml:"lock_path,omitempty"`
	UploadTmpDir string `mapstructure:"upload_tmpdir" json:"upload_tmpdir,omitempty" yaml:"upload_tmpdir,omitempty"`

	HealthCheckInterval int `mapstructure:"healthcheck_interval" json:"healthcheck_interval,omitempty" yaml:"healthcheck_interval,omitempty"`
	TaskDeadline        int `mapstructure:"task_deadline" json:"task_deadline,omitempty" yaml:"task_deadline,omitempty"`
	NodeDeadline        int `mapstructure:"node_deadline" json:"node_deadline,omitempty" yaml:"node_deadline,omitempty"`

	SessionProvider       string `mapstructure:"session_provider" json:"session_provider,omitempty" yaml:"session_provider,omitempty"`
	SessionProviderConfig string `mapstructure:"session_provider_config" json:"session_provider_config,omitempty" yaml:"session_provider_config,omitempty"`

	//Pagination
	MaxPageSize int `mapstructure:"max_page_size" json:"max_page_size,omitempty" yaml:"max_page_size,omitempty"`
}

type StorageConfig struct {
	Type string `mapstructure:"type" json:"type" yaml:"type"`

	ArtefactPath  string `mapstructure:"artefact_path" json:"artefact_path,omitempty" yaml:"artefact_path,omitempty"`
	NamespacePath string `mapstructure:"namespace_path" json:"namespace_path,omitempty" yaml:"namespace_path,omitempty"`
	StoragePath   string `mapstructure:"storage_path" json:"storage_path,omitempty" yaml:"storage_path,omitempty"`
}

type DatabaseConfig struct {
	DBEngine string `mapstructure:"engine" json:"engine" yaml:"engine"`
	DBPath   string `mapstructure:"db_path" json:"db_path,omitempty" yaml:"db_path,omitempty"`

	Endpoints    []string `mapstructure:"db_endpoints" json:"db_endpoints,omitempty" yaml:"db_endpoints,omitempty"`
	User         string   `mapstructure:"db_user" json:"db_user,omitempty" yaml:"db_user,omitempty"`
	DatabaseName string   `mapstructure:"db_name" json:"db_name,omitempty" yaml:"db_name,omitempty"`
	Password     string   `mapstructure:"db_password" json:"db_password,omitempty" yaml:"db_password,omitempty"`
	CertPath     string   `mapstructure:"db_certpath" json:"db_certpath,omitempty" yaml:"db_certpath,omitempty"`
	KeyPath      string   `mapstructure:"db_keypath" json:"db_keypath,omitempty" yaml:"db_keypath,omitempty"`

	// Tiedot configs settings
	TiedotDocMaxRoom    int  `mapstructure:"tiedot_doc_maxroom" json:"tiedot_doc_maxroom,omitempty" yaml:"tiedot_doc_maxroom,omitempty"`          // DocMaxRoom is the maximum size of a single document that will ever be accepted into database.
	TiedotColFileGrowth int  `mapstructure:"tiedot_colfile_growth" json:"tiedot_colfile_growth,omitempty" yaml:"tiedot_colfile_growth,omitempty"` // ColFileGrowth is the size (in bytes) to grow collection data file when new documents have to fit in.
	TiedotPerBucket     int  `mapstructure:"tiedot_per_bucket" json:"tiedot_per_bucket,omitempty" yaml:"tiedot_per_bucket,omitempty"`             // PerBucket is the number of entries pre-allocated to each hash table bucket.
	TiedotHTFileGrowth  int  `mapstructure:"tiedot_htfilegrowth" json:"tiedot_htfilegrowth,omitempty" yaml:"tiedot_htfilegrowth,omitempty"`       /// HTFileGrowth is the size (in bytes) to grow hash table file to fit in more entries.
	TiedotHashBits      uint `mapstructure:"tiedot_hashbits" json:"tiedot_hashbits,omitempty" yaml:"tiedot_hashbits,omitempty"`                   // HashBits is the number of bits to consider for hashing indexed key, also determines the initial number of buckets in a hash table file.
}

type AgentConfig struct {
	SecretKey          string         `mapstructure:"secret_key" json:"secret_key,omitempty" yaml:"secret_key,omitempty"`
	BuildPath          string         `mapstructure:"build_path" json:"build_path" yaml:"build_path"`
	AgentConcurrency   int            `mapstructure:"concurrency" json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	AgentKey           string         `mapstructure:"agent_key" json:"agent_key" yaml:"agent_key"`
	ApiKey             string         `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
	PrivateQueue       int            `mapstructure:"private_queue" json:"private_queue,omitempty" yaml:"private_queue,omitempty"`
	StandAlone         bool           `mapstructure:"standalone" json:"standalone,omitempty" yaml:"standalone,omitempty"`
	DownloadRateLimit  int64          `mapstructure:"download_speed_limit" json:"download_speed_limit,omitempty" yaml:"download_speed_limit,omitempty"`
	UploadRateLimit    int64          `mapstructure:"upload_speed_limit" json:"upload_speed_limit,omitempty" yaml:"upload_speed_limit,omitempty"`
	Queues             map[string]int `mapstructure:"queues" json:"queues,omitempty" yaml:"queues,omitempty"`
	UploadChunkSize    int            `mapstructure:"upload_chunk_size" json:"upload_chunk_size,omitempty" yaml:"upload_chunk_size,omitempty"`
	SupportedExecutors []string       `mapstructure:"executor" json:"executor,omitempty" yaml:"executor,omitempty"`
	ForceAgentId       string         `mapstructure:"force_agent_id" json:"force_agent_id,omitempty" yaml:"force_agent_id,omitempty"`

	// List of command to execute before execute a task
	PreTaskHookExec []string `mapstructure:"pre_task_hook_exec" json:"pre_task_hook_exec,omitempty" yaml:"pre_task_hook_exec,omitempty"`

	DockerEndpoint    string   `mapstructure:"docker_endpoint" json:"docker_endpoint,omitempty" yaml:"docker_endpoint,omitempty"`
	DockerKeepImg     bool     `mapstructure:"docker_keepimg" json:"docker_keepimg,omitempty" yaml:"docker_keepimg,omitempty"`
	DockerPriviledged bool     `mapstructure:"docker_privileged" json:"docker_privileged,omitempty" yaml:"docker_privileged,omitempty"`
	DockerInDocker    bool     `mapstructure:"docker_in_docker" json:"docker_in_docker,omitempty" yaml:"docker_in_docker,omitempty"`
	DockerEndpointDiD string   `mapstructure:"docker_in_docker_endpoint" json:"docker_in_docker_endpoint,omitempty" yaml:"docker_in_docker_endpoint,omitempty"`
	DockerCaps        []string `mapstructure:"docker_caps" json:"docker_caps,omitempty" yaml:"docker_caps,omitempty"`
	DockerCapsDrop    []string `mapstructure:"docker_caps_drop" json:"docker_caps_drop,omitempty" yaml:"docker_caps_drop,omitempty"`
	DefaultTaskQuota  string   `mapstructure:"default_task_quota" json:"default_task_quota,omitempty" yaml:"default_task_quota,omitempty"`

	KubeConfigPath   string `mapstructure:"kubeconfig" json:"kubeconfig,omitempty" yaml:"kubeconfig,omitempty"`
	KubeNamespace    string `mapstructure:"kube_namespace" json:"kube_namespace,omitempty" yaml:"kube_namespace,omitempty"`
	KubeStorageClass string `mapstructure:"kube_storageclass" json:"kube_storageclass,omitempty" yaml:"kube_storageclass,omitempty"`
	KubeDropletImage string `mapstructure:"kube_droplet_image" json:"kube_droplet_image,omitempty" yaml:"kube_droplet_image,omitempty"`

	LxdEndpoint            string            `mapstructure:"lxd_endpoint" json:"lxd_endpoint,omitempty" yaml:"lxd_endpoint,omitempty"`
	LxdConfigDir           string            `mapstructure:"lxd_config_dir" json:"lxd_config_dir,omitempty" yaml:"lxd_config_dir,omitempty"`
	LxdDisableLocal        bool              `mapstructure:"lxd_disable_local" json:"lxd_disable_local,omitempty" yaml:"lxd_disable_local,omitempty"`
	LxdProfiles            []string          `mapstructure:"lxd_profiles" json:"lxd_profiles,omitempty" yaml:"lxd_profiles,omitempty"`
	LxdEphemeralContainers bool              `mapstructure:"lxd_ephemeral_containers" json:"lxd_ephemeral_containers,omitempty" yaml:"lxd_ephemeral_containers,omitempty"`
	LxdCacheRegistry       map[string]string `mapstructure:"lxd_cache_registry" json:"lxd_cache_registry,omitempty" yaml:"lxd_cache_registry,omitempty"`

	CacheRegistryCredentials map[string]string `mapstructure:"cache_registry" json:"cache_registry,omitempty" yaml:"cache_registry,omitempty"`

	HealthCheckExec      []string `mapstructure:"health_check_exec" json:"health_check_exec,omitempty" yaml:"health_check_exec,omitempty"`
	HealthCheckCleanPath []string `mapstructure:"health_check_clean_path" json:"health_check_clean_path,omitempty" yaml:"health_check_clean_path,omitempty"`
}

type GeneralConfig struct {
	Debug         bool   `mapstructure:"debug" json:"debug,omitempty" yaml:"debug,omitempty"`
	LogFile       string `mapstructure:"logfile" json:"logfile,omitempty" yaml:"logfile,omitempty"`
	LogLevel      string `mapstructure:"loglevel" json:"loglevel,omitempty" yaml:"loglevel,omitempty" `
	TLSCert       string `mapstructure:"tls_cert" json:"tls_cert,omitempty" yaml:"tls_cert,omitempty"`
	TLSKey        string `mapstructure:"tls_key" json:"tls_key,omitempty" yaml:"tls_key,omitempty"`
	ClientTimeout int    `mapstructure:"client_timeout" json:"client_timeout,omitempty" yaml:"client_timeout,omitempty"`
}

type Config struct {
	Viper *v.Viper `json:"-" yaml:"-"`

	General  GeneralConfig  `mapstructure:"general" json:"general,omitempty" yaml:"general,omitempty"`
	Web      WebConfig      `mapstructure:"web" json:"web,omitempty" yaml:"web,omitempty"`
	Storage  StorageConfig  `mapstructure:"storage" json:"storage,omitempty" yaml:"storage,omitempty"`
	Database DatabaseConfig `mapstructure:"db" json:"db,omitempty" yaml:"db,omitempty"`
	Agent    AgentConfig    `mapstructure:"agent" json:"agent,omitempty" yaml:"agent,omitempty"`
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
	viper.SetDefault("web.max_page_size", 300)

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

	// Tiedot default options
	viper.SetDefault("db.tiedot_doc_maxroom", 2*1048576)
	//	viper.SetDefault("db.tiedot_colfile_growth", 32*1048576)
	viper.SetDefault("db.tiedot_colfile_growth", 0.5*1048576)
	viper.SetDefault("db.tiedot_per_bucket", 1)
	viper.SetDefault("db.tiedot_htfilegrowth", 0.5*1048576)
	//viper.SetDefault("db.tiedot_htfilegrowth", 32*1048576)
	viper.SetDefault("db.tiedot_hashbits", 16)

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
	viper.SetDefault("agent.lxd_disable_local", false)
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
	data, _ := yaml.Marshal(c)
	return string(data)
}

func (c *GeneralConfig) String() string {
	data, _ := yaml.Marshal(c)
	return string(data)
}

func (c *AgentConfig) String() string {
	data, _ := yaml.Marshal(c)
	return string(data)
}

func (c *Config) String() string {
	data, _ := c.Yaml()
	return string(data)
}

func (c *Config) Yaml() ([]byte, error) {
	return yaml.Marshal(c)
}

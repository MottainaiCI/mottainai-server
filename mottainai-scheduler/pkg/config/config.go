/*

Copyright (C) 2018-2021  Daniele Rondina <geaaru@sabayonlinux.org>

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

package config

import (
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	v "github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	MOTTAINAI_SCHED_ENV_PREFIX = "MOTTAINAI_SCHED"
	MOTTAINAI_SCHED_CONFIGNAME = "mottainai-scheduler"
)

type Config struct {
	Viper *v.Viper

	General   setting.GeneralConfig `mapstructure:"general" json:"general" yaml:"general"`
	Scheduler SchedulerConfig       `mapstructure:"scheduler json:"scheduler" yaml:"scheduler""`
	Web       WebConfig             `mapstructure:"web" json:"web" yaml:"web"`
}

type SchedulerConfig struct {
	ApiKey string   `mapstructure:"api_key" json:"api_key,omitempty" yaml:"api_key,omitempty"`
	Queues []string `mapstructure:"queues" json:"queues,omitempty" yaml:"queues,omitempty"`

	ScheduleTimerSec    int `mapstructure:"schedule_timer_sec,omitempty" json:"schedule_timer_sec,omitempty" yaml:"schedule_timer_sec,omitempty"`
	AgentDeadTimeoutSec int `mapstructure:"agent_dead_timeout,omitempty json:"agent_dead_timeout,omitempty" yaml:"schedule_timer_sec,omitempty"`
}

type WebConfig struct {
	// TODO: TO rename in API URL
	AppURL string `mapstructure:"application_url" json:"application_url" yaml:"application_url"`
}

func NewConfig(viper *v.Viper) *Config {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &Config{Viper: viper}
}

func (c *Config) GetGeneral() *setting.GeneralConfig {
	return &c.General
}

func (c *Config) GetScheduler() *SchedulerConfig {
	return &c.Scheduler
}

func (c *Config) GetWeb() *WebConfig {
	return &c.Web
}

func (c *Config) GenDefault() {
	GenDefault(c.Viper)
}

func GenDefault(viper *v.Viper) {

	viper.SetDefault("web.application_url", "http://127.0.0.1:9090")

	viper.SetDefault("scheduler.api_key", "")
	viper.SetDefault("scheduler.queues", []string{})
	viper.SetDefault("scheduler.schedule_timer_sec", 5)
	viper.SetDefault("scheduler.agent_dead_timeout", 1200)

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

func (c *Config) String() string {
	data, _ := c.Yaml()
	return string(data)
}

func (c *Config) Yaml() ([]byte, error) {
	return yaml.Marshal(c)
}

func (c *Config) ToMottainaiConfig() *setting.Config {
	cfg := setting.NewConfig(c.Viper)
	cfg.General = c.General
	cfg.GetWeb().AppURL = c.Web.AppURL

	return cfg
}

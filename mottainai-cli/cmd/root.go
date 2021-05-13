/*

Copyright (C) 2017-2021  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>

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

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	namespace "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/namespace"
	node "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/node"
	nodequeuecmd "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/nodequeue"
	pipeline "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/pipeline"
	plan "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/plan"
	profile "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/profile"
	queuecmd "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/queue"
	secret "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/secret"
	settingcmd "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/settings"
	webhookcmd "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/webhook"

	debug "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/debug"
	simulate "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/simulate"
	storage "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/storage"
	task "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/task"
	token "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/token"
	user "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/user"

	common "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	viper "github.com/spf13/viper"
)

const (
	cliExamples = `$> mottainai-cli -m http://127.0.0.1:8080 task create --json task.json

$> mottainai-cli -m http://127.0.0.1:8080 namespace list
`
)

func initConfig(config *setting.Config) {
	// Set env variable
	config.Viper.SetEnvPrefix(common.MCLI_ENV_PREFIX)
	config.Viper.BindEnv("config")
	config.Viper.SetDefault("master", "http://localhost:8080")
	config.Viper.SetDefault("profile", "")
	config.Viper.SetDefault("config", "")
	config.Viper.SetDefault("etcd-config", false)

	config.Viper.AutomaticEnv()

	// Set config file name (without extension)
	config.Viper.SetConfigName(common.MCLI_CONFIG_NAME)

	// Set Config paths list
	config.Viper.AddConfigPath(common.MCLI_LOCAL_PATH)
	config.Viper.AddConfigPath(
		fmt.Sprintf("%s/%s", common.GetHomeDir(), common.MCLI_HOME_PATH))

	// Create EnvKey Replacer for handle complex structure
	replacer := strings.NewReplacer(".", "__")
	config.Viper.SetEnvKeyReplacer(replacer)
	config.Viper.SetTypeByDefaultValue(true)
}

func initCommand(rootCmd *cobra.Command, config *setting.Config) {
	var pflags = rootCmd.PersistentFlags()
	v := config.Viper

	pflags.StringP("master", "m", "http://localhost:8080", "MottainaiCI webUI URL")
	pflags.StringP("apikey", "k", "fb4h3bhgv4421355", "Mottainai API key")

	pflags.StringP("profile", "p", "", "Use specific profile for call API.")

	v.BindPFlag("master", rootCmd.PersistentFlags().Lookup("master"))
	v.BindPFlag("apikey", rootCmd.PersistentFlags().Lookup("apikey"))
	v.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))

	rootCmd.AddCommand(
		task.NewTaskCommand(config),
		node.NewNodeCommand(config),
		token.NewTokenCommand(config),
		namespace.NewNamespaceCommand(config),
		plan.NewPlanCommand(config),
		profile.NewProfileCommand(config),
		user.NewUserCommand(config),
		storage.NewStorageCommand(config),
		simulate.NewSimulateCommand(config),
		pipeline.NewPipelineCommand(config),
		settingcmd.NewSettingCommand(config),
		webhookcmd.NewWebHookCommand(config),
		secret.NewSecretCommand(config),
		debug.NewDebugCommand(config),
		queuecmd.NewQueueCommand(config),
		nodequeuecmd.NewNodeQueueCommand(config),
	)
}

func Execute() {
	// Create Main Instance Config object
	var config *setting.Config = setting.NewConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        common.MCLI_HEADER,
		Version:      setting.MOTTAINAI_VERSION,
		Example:      cliExamples,
		Args:         cobra.OnlyValidArgs,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			var v *viper.Viper = config.Viper

			// Parse configuration file
			err = config.Unmarshal()
			// TODO: Add loglevel in debug that said no config file processed.
			// if err != nil {
			//	fmt.Println(err)
			//}

			// Load profile data and override master if not present.
			if v.Get("profiles") != nil && !cmd.Flag("master").Changed {

				// PRE: profiles contains a map
				//      map[
				//        <NAME_PROFILE1>:<PROFILE INTERFACE>
				//        <NAME_PROFILE2>:<PROFILE INTERFACE>
				//     ]

				var conf common.ProfileConf
				var profile *common.Profile
				if err = v.Unmarshal(&conf); err != nil {
					fmt.Println("Ignore config: ", err)
				} else {
					if v.GetString("profile") != "" {
						profile, err = conf.GetProfile(v.GetString("profile"))

						if profile != nil {
							v.Set("master", profile.GetMaster())
							if profile.GetApiKey() != "" && !cmd.Flag("apikey").Changed {
								v.Set("apikey", profile.GetApiKey())
							}
						} else {
							fmt.Printf("No profile with name %s. I use default value.\n", v.GetString("profile"))
						}
					}
				}

			}
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

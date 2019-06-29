/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package mottainai

import (
	"os"
	"path"
	"strings"
	"sync"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	taskmanager "github.com/MottainaiCI/mottainai-server/pkg/tasks/manager"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	"github.com/RichardKnop/machinery/v1/log"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
)

func (m *MottainaiAgent) HealthCheckSetup(force bool) {
	m.Invoke(func(config *setting.Config) {
		th := taskmanager.DefaultTaskHandler(config)
		m.Map(th)
		fetcher := client.NewClient(config.GetWeb().AppURL, config)
		fetcher.SetToken(config.GetAgent().ApiKey)
		m.Map(fetcher)
		m.TimerSeconds(int64(800), true, func() { m.HealthClean(force) })
	})
}

func (m *MottainaiAgent) AgentIsBusy() bool {
	var busy bool = false
	m.Invoke(func(c *client.Fetcher, config *setting.Config) {
		var tlist []agenttasks.Task

		err := c.NodesTask(config.GetAgent().AgentKey, &tlist)
		if err != nil {
			log.ERROR.Println("> Error getting task running on this host - skipping deep host cleanup")
			busy = true
		}
		for _, t := range tlist {
			if t.IsRunning() {
				log.INFO.Println("> Task running on the host, skipping deep host cleanup")
				busy = true
			}
		}

	})

	return busy
}

func (m *MottainaiAgent) HealthCheckRun(force bool) {
	m.HealthCheckSetup(force)
	m.Anagent.Start()
}

func (m *MottainaiAgent) HealthClean(force bool) {
	m.CleanBuildDir(force)

	m.Invoke(func(c *client.Fetcher, config *setting.Config) {

		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			m.CleanHealthCheckExec()
		}()
		go func() {
			defer wg.Done()
			m.CleanHealthCheckPathHost()
		}()
		log.INFO.Println("> Waiting for cleanup operations to end")
		wg.Wait()
		log.INFO.Println("> Done")
	})
}

// FIXME: temp (racy) workaround
// As vagrant does not guarantee removal of imported boxes, cleanup periodically
func (m *MottainaiAgent) CleanHealthCheckPathHost() {

	m.Invoke(func(config *setting.Config) {
		for _, k := range config.GetAgent().HealthCheckCleanPath {
			log.INFO.Println("> Removing dangling files in " + k)
			if err := utils.RemoveContents(k); err != nil {
				log.ERROR.Println("> Failed removing contents from ", k, " ", err.Error())
			}
		}
	})
}

func (m *MottainaiAgent) CleanHealthCheckExec() {
	m.Invoke(func(config *setting.Config) {
		for _, k := range config.GetAgent().HealthCheckExec {
			log.INFO.Println("> Executing: " + k)
			args := strings.Split(k, " ")
			cmdName := args[0]
			out, stderr, err := utils.Cmd(cmdName, args[1:])
			if err != nil {
				log.ERROR.Println("!! Error: ", err.Error()+": "+stderr)
			}
			log.INFO.Println(out)
		}
	})
}

func (m *MottainaiAgent) IsAgentBusyWith(id string) bool {
	var busy bool = true
	m.Invoke(func(c *client.Fetcher, config *setting.Config) {
		c.Doc(id)
		th := taskmanager.DefaultTaskHandler(config)
		task_info := th.FetchTask(c)
		if th.Err != nil {
			log.INFO.Println("Error fetching task: " + th.Err.Error())
			return
		}
		if task_info.IsDone() || task_info.ID == "" {
			busy = false
		}
	})
	return busy
}

func (m *MottainaiAgent) CleanBuildDir(force bool) {
	m.Invoke(func(config *setting.Config) {
		log.INFO.Println("Cleaning " + config.GetAgent().BuildPath)

		stuff, err := utils.ListAll(config.GetAgent().BuildPath)
		if err != nil {
			panic(err)
		}

		defer func() {
			if r := recover(); r != nil {
				log.ERROR.Println(r)
			}
		}()

		for _, what := range stuff {
			log.INFO.Println("Found: " + what)

			if force || !m.IsAgentBusyWith(what) {
				if what == "lxc" {
					log.INFO.Println("Keeping: " + what)
					continue
				}

				log.INFO.Println("Removing: " + what)
				os.RemoveAll(path.Join(config.GetAgent().BuildPath, what))
			} else {
				log.INFO.Println("Keeping: " + what)
			}
		}

	})
}

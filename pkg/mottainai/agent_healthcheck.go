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

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	"github.com/RichardKnop/machinery/v1/log"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
)

func (m *MottainaiAgent) HealthCheckSetup(config string) {
	setting.GenDefault()
	if len(config) > 0 {
		setting.LoadFromFileEnvironment(config)
	}

	th := agenttasks.DefaultTaskHandler()
	m.Map(th)
	ID := utils.GenID()
	hostname := utils.Hostname()
	log.INFO.Println("Worker ID: " + ID)
	log.INFO.Println("Worker Hostname: " + hostname)

	fetcher := client.NewClient()
	fetcher.RegisterNode(ID, hostname)
	m.Map(fetcher)

	m.TimerSeconds(int64(800), true, func() { m.HealthClean() })
}

func (m *MottainaiAgent) HealthClean() {
	m.CleanBuildDir()
}

func (m *MottainaiAgent) CleanBuildDir() {
	m.Invoke(func(c *client.Fetcher) {
		log.INFO.Println("Cleaning " + setting.Configuration.BuildPath)

		stuff, err := utils.ListAll(setting.Configuration.BuildPath)
		if err != nil {
			panic(err)
		}

		defer func() {
			if r := recover(); r != nil {
				log.ERROR.Println(r)
			}
		}()

		for _, what := range stuff {
			c.Doc(what)
			th := agenttasks.DefaultTaskHandler()
			task_info := th.FetchTask(c)
			log.INFO.Println("Found: " + what)
			log.INFO.Println(task_info)
			if task_info.IsDone() {
				log.INFO.Println("Removing: " + what)
				os.Remove(path.Join(setting.Configuration.BuildPath, what))
			} else {
				log.INFO.Println("Keeping: " + what)
			}
		}

	})
}

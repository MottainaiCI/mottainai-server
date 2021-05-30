/*

Copyright (C) 2021  Daniele Rondina <geaaru@sabayonlinux.org>

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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/entities"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	"github.com/mudler/anagent"
)

type MottainaiExporter struct {
	*anagent.Anagent
}

func NewExporter(config *setting.Config) *MottainaiExporter {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			fmt.Println("Received SIGINT event. Shutdown.")
		case syscall.SIGTERM:
			fmt.Println("Received SIGTERM event. Shutdown.")
		}

		os.Exit(0)
	}()

	ans := &MottainaiExporter{Anagent: anagent.New()}

	ans.Map(config)

	database.NewDatabase(config)
	ans.Map(database.DBInstance)

	return ans
}

func (m *MottainaiExporter) Export(targetDir string, withDateDir bool) error {
	var err error

	now := time.Now().UTC()

	if targetDir == "" {
		return errors.New("Invalid target dir")
	}

	if withDateDir {
		targetDir = path.Join(targetDir,
			fmt.Sprintf("%4d%02d%02d_%02d:%02d",
				now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()),
		)
	}

	if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}

	m.Invoke(func(d *database.Database, config *setting.Config) {

		for _, c := range entities.GetMottainaiEntities() {

			targetFile := filepath.Join(targetDir, fmt.Sprintf("%s.json", c))
			var data []byte

			// TODO, converter in memory marshal to reader/writer.
			switch c {
			case entities.Webhooks:
				data, err = json.Marshal(d.Driver.AllWebHooks())
			case entities.Tasks:
				data, err = json.Marshal(d.Driver.AllTasks(config))
			case entities.Secrets:
				data, err = json.Marshal(d.Driver.AllSecrets())
			case entities.Users:
				data, err = json.Marshal(d.Driver.AllUsers())
			case entities.Plans:
				data, err = json.Marshal(d.Driver.AllPlans(config))
			case entities.Pipelines:
				data, err = json.Marshal(d.Driver.AllPipelines(config))
			case entities.Nodes:
				data, err = json.Marshal(d.Driver.AllNodes())
			case entities.Namespaces:
				data, err = json.Marshal(d.Driver.AllNamespaces())
			case entities.Tokens:
				data, err = json.Marshal(d.Driver.AllTokens())
			case entities.Artefacts:
				data, err = json.Marshal(d.Driver.AllStorages())
			case entities.Organizations:
				data, err = json.Marshal(d.Driver.AllOrganizations())
			case entities.Settings:
				data, err = json.Marshal(d.Driver.AllSettings())
			default:
				// POST: ignoring queues and nodeques. They are created automatically.
				continue
			}

			if err != nil {
				fmt.Println(fmt.Sprintf(
					"Error on marshal %s entities: %s",
					c, err.Error()))
				return
			}

			err = ioutil.WriteFile(targetFile, data, 0640)
			if err != nil {
				fmt.Println(fmt.Sprintf(
					"Error on write file %s: %s",
					targetFile, err.Error()))
				return
			}

			fmt.Println(fmt.Sprintf("Exported entity %s to file %s.",
				c, targetFile))

		}
	})

	return err
}

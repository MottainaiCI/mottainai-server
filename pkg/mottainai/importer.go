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
	"path/filepath"
	"syscall"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/entities"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"

	"github.com/mudler/anagent"
)

type MottainaiImporter struct {
	*anagent.Anagent
}

func NewImporter(config *setting.Config) *MottainaiImporter {
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

	ans := &MottainaiImporter{Anagent: anagent.New()}

	ans.Map(config)

	database.NewDatabase(config)
	ans.Map(database.DBInstance)

	return ans
}

func (m *MottainaiImporter) Import(backupDir string) error {
	var err error

	if backupDir == "" {
		return errors.New("Invalid dir")
	}

	m.Invoke(func(d *database.Database, config *setting.Config) {

		var data []byte

		for _, c := range entities.GetMottainaiEntities() {

			// skip queue and node queues
			if c == "queues" || c == "nodequeues" {
				continue
			}

			inputFile := filepath.Join(backupDir, fmt.Sprintf("%s.json", c))

			existsFile, err := utils.Exists(inputFile)
			if err != nil {
				fmt.Println(fmt.Sprintf(
					"Error on check if exists the file %s. Skipping entity %s.",
					inputFile, c,
				))
				continue
			}

			if !existsFile {
				fmt.Println(fmt.Sprintf(
					"File %s not found. Skipping entity %s.",
					inputFile, c,
				))
				continue
			}

			data, err = ioutil.ReadFile(inputFile)
			if err != nil {
				err = errors.New(
					fmt.Sprintf("Error on read file %s: %s",
						inputFile, err.Error()))
				return
			}

			fmt.Println(fmt.Sprintf("Importing %s...", c))

			objects := make([]map[string]interface{}, 0)
			err = json.Unmarshal(data, &objects)
			if err != nil {
				return
			}

			// TODO, converter in memory marshal to reader/writer.
			switch c {
			case entities.Webhooks:
				for _, w := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), w["id"].(string), w)
					if err != nil {
						return
					}
				}
			case entities.Tasks:
				for _, t := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), t["ID"].(string), t)
					if err != nil {
						return
					}
				}
			case entities.Secrets:
				for _, s := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), s["id"].(string), s)
					if err != nil {
						return
					}
				}
			case entities.Users:
				for _, u := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), u["id"].(string), u)
					if err != nil {
						return
					}
				}
			case entities.Plans:
				for _, p := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), p["ID"].(string), p)
					if err != nil {
						return
					}
				}
			case entities.Pipelines:
				for _, p := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), p["ID"].(string), p)
					if err != nil {
						return
					}
				}
			case entities.Nodes:
				for _, n := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), n["ID"].(string), n)
					if err != nil {
						return
					}
				}
			case entities.Namespaces:
				for _, n := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), n["ID"].(string), n)
					if err != nil {
						return
					}
				}
			case entities.Tokens:
				for _, t := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), t["id"].(string), t)
					if err != nil {
						return
					}
				}
			case entities.Artefacts:
				for _, a := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), a["ID"].(string), a)
					if err != nil {
						return
					}
				}
			case entities.Organizations:
				for _, o := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), o["id"].(string), o)
					if err != nil {
						return
					}
				}
			case entities.Settings:
				for _, s := range objects {
					err = d.Driver.RestoreDoc(d.Driver.GetCollectionName(c), s["id"].(string), s)
					if err != nil {
						return
					}
				}

			default:
				// POST: ignoring queues and nodeques. They are created automatically.
				continue
			}

			fmt.Println(fmt.Sprintf("Loaded %d %s.", len(objects), c.String()))
		}
	})

	return err
}

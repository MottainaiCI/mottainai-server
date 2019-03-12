/*

Copyright (C) 2017-2019  Ettore Di Giacinto <mudler@gentoo.org>
Some code portions and re-implemented design are also coming
from the Gogs project, which is using the go-macaron framework and was
really source of ispiration. Kudos to them!

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

package client_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	. "github.com/MottainaiCI/mottainai-server/pkg/client"
	helpers "github.com/MottainaiCI/mottainai-server/tests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// a little code dup for coverage (see tests/helpers)
func NewFakeClient() (*Fetcher, error) {
	if len(helpers.Tokens) == 0 {
		return nil, errors.New("No tokens registered in the helper")
	}

	return NewTokenClient(helpers.Config.GetWeb().AppURL, helpers.Tokens[0].Key, helpers.Config), nil
}

func OSReadDir(root string) ([]string, error) {
	var files []string
	f, err := os.Open(root)
	if err != nil {
		return files, err
	}
	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return files, err
	}
	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}

var _ = Describe("Client", func() {
	dir, err := ioutil.TempDir("", "client_test")
	defer os.RemoveAll(dir) // clean up
	errserver := helpers.StartServer(dir)

	Describe("Client download", func() {
		Context("Default fixture ", func() {
			It("Get all the task artefacts", func() {

				Expect(err).ToNot(HaveOccurred())
				Expect(errserver).ToNot(HaveOccurred())

				download, err := ioutil.TempDir("", "client_download")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(download) // clean up

				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())

				fixtures := []string{"fixture1#test.vvvsz", "fixture1", "fixture2", "fixture3.sh", "b000gy.@bogyy"}

				fixture1 := filepath.Join(helpers.Config.GetStorage().ArtefactPath, helpers.Tasks[0], "simple")
				os.MkdirAll(fixture1, os.ModePerm)

				for _, f := range fixtures {
					helpers.CreateFile(10, filepath.Join(fixture1, f))
				}

				// Create a fake artefact for task id 11111
				fixture2 := filepath.Join(helpers.Config.GetStorage().ArtefactPath, helpers.Tasks[0], "#test:")
				os.MkdirAll(fixture2, os.ModePerm)
				for _, f := range fixtures {
					helpers.CreateFile(10, filepath.Join(fixture2, f))
				}

				fetcher.DownloadArtefactsFromTask(helpers.Tasks[0], download)

				for _, f := range fixtures {
					_, err = os.Stat(filepath.Join(download, "simple", f))
					Expect(err).ToNot(HaveOccurred())
					_, err = os.Stat(filepath.Join(download, "#test:", f))
					Expect(err).ToNot(HaveOccurred())
				}

			})
		})
	})

	Describe("Task report", func() {
		Context("Default task", func() {
			It("Reports the output correctly", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])
				_, err = fetcher.AppendTaskOutput("test")
				Expect(err).ToNot(HaveOccurred())
				_, err = fetcher.AppendTaskOutput("foobar")
				Expect(err).ToNot(HaveOccurred())

				logfile := filepath.Join(helpers.Config.GetStorage().ArtefactPath, helpers.Tasks[0], "build_"+helpers.Tasks[0]+".log")
				_, err = os.Stat(logfile)
				Expect(err).ToNot(HaveOccurred())

				dat, err := ioutil.ReadFile(logfile)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(dat)).Should(ContainSubstring("test"))
				Expect(string(dat)).Should(ContainSubstring("foobar"))
			})
		})

		Context("Executor Report", func() {
			It("Reports the output correctly", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				e := tasks.TaskExecutor{MottainaiClient: fetcher, Context: tasks.NewExecutorContext(), Config: helpers.Config}
				e.Report("hunk1", "mottainai")

				logfile := filepath.Join(helpers.Config.GetStorage().ArtefactPath, helpers.Tasks[0], "build_"+helpers.Tasks[0]+".log")
				_, err = os.Stat(logfile)
				Expect(err).ToNot(HaveOccurred())

				dat, err := ioutil.ReadFile(logfile)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(dat)).Should(ContainSubstring("hunk1"))
				Expect(string(dat)).Should(ContainSubstring("mottainai"))
			})
		})

	})

})

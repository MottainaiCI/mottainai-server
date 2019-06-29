/*

Copyright (C) 2019  Ettore Di Giacinto <mudler@gentoo.org>

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
	"strings"

	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	"github.com/MottainaiCI/mottainai-server/pkg/secret"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	executors "github.com/MottainaiCI/mottainai-server/pkg/tasks/executors"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	. "github.com/MottainaiCI/mottainai-server/pkg/client"
	helpers "github.com/MottainaiCI/mottainai-server/tests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// a little code dup for coverage (see tests/helpers)
func NewFakeClient() (HttpClient, error) {
	if len(helpers.Tokens) == 0 {
		return nil, errors.New("No tokens registered in the helper")
	}

	return NewTokenClient(helpers.Config.GetWeb().AppURL, helpers.Tokens[0].Key, helpers.Config), nil
}

func ExpectSuccessfulResponse(ev event.APIResponse, err error) {
	Expect(err).ToNot(HaveOccurred())
	Expect(ev.Error).To(Equal(""))
	Expect(ev.Processed).To(Equal("true"))
	Expect(ev.Status).To(Equal("ok"))
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
	fixtures := []string{"fixture1#test.vvvsz", "fixture1", "fixture2", "fixture3.sh", "b000gy.@bogyy"}

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
				fetcher.Doc(helpers.Tasks[0])

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

				testfile := filepath.Join(dir, "testfffffff")
				helpers.CreateFile(10, testfile)
				err = fetcher.UploadArtefact(testfile, "/")
				Expect(err).ToNot(HaveOccurred())

				list, err := fetcher.TaskFileList(helpers.Tasks[0])
				Expect(err).ToNot(HaveOccurred())
				Expect(strings.Join(list, " ")).Should(ContainSubstring("/testfffffff"))
				Expect(strings.Join(list, " ")).Should(ContainSubstring("/simple/fixture1#test.vvvsz"))
			})
		})
	})

	Describe("Task Report", func() {
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

				e := executors.TaskExecutor{MottainaiClient: fetcher, Context: executors.NewExecutorContext(), Config: helpers.Config}
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

	Describe("Client API calls", func() {

		Context("Namespace", func() {
			It("Can tag and download artefacts", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				download, err := ioutil.TempDir("", "client_download")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(download) // clean up

				download2, err := ioutil.TempDir("", "client_download2")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(download2) // clean up

				ev, err := fetcher.NamespaceCreate("test")
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.NamespaceTag(helpers.Tasks[0], "test2")
				ExpectSuccessfulResponse(ev, err)

				fetcher.DownloadArtefactsFromNamespace("test2", download)
				for _, f := range fixtures {
					_, err = os.Stat(filepath.Join(download, "simple", f))
					Expect(err).ToNot(HaveOccurred())
					_, err = os.Stat(filepath.Join(download, "#test:", f))
					Expect(err).ToNot(HaveOccurred())
				}

				ev, err = fetcher.NamespaceRemovePath("test2", "simple/fixture1")
				ExpectSuccessfulResponse(ev, err)

				fetcher.DownloadArtefactsFromNamespace("test2", download2)
				_, err = os.Stat(filepath.Join(download2, "simple", "fixture1"))
				Expect(err).To(HaveOccurred())

				for _, f := range []string{"fixture1#test.vvvsz", "fixture2", "fixture3.sh", "b000gy.@bogyy"} {
					_, err = os.Stat(filepath.Join(download2, "simple", f))
					Expect(err).ToNot(HaveOccurred())
					_, err = os.Stat(filepath.Join(download2, "#test:", f))
					Expect(err).ToNot(HaveOccurred())
				}

				ev, err = fetcher.NamespaceDelete("test2")
				ExpectSuccessfulResponse(ev, err)

			})
		})

		Context("Setting", func() {
			It("Can create and update them", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				ev, err := fetcher.SettingCreate(map[string]interface{}{"key": "test", "value": "foo"})
				Expect(err).ToNot(HaveOccurred())
				Expect(ev.Processed).To(Equal("true"))
				Expect(ev.Status).To(Equal("ok"))

				ev, err = fetcher.SettingUpdate(map[string]interface{}{"key": "test", "value": "a"})
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.SettingRemove("test")
				ExpectSuccessfulResponse(ev, err)

			})
		})

		Context("Node", func() {
			It("Can create, update and remove them", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				ev, err := fetcher.CreateNode()
				id := ev.ID
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.RemoveNode(id)
				ExpectSuccessfulResponse(ev, err)

			})
		})

		Context("Storage", func() {
			It("Can create and update them", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				ev, err := fetcher.StorageCreate("teststorage1")
				id := ev.ID
				ExpectSuccessfulResponse(ev, err)

				testfile := filepath.Join(dir, "test")
				helpers.CreateFile(10, testfile)
				sum, err := helpers.FileSum(testfile)
				Expect(err).ToNot(HaveOccurred())

				err = fetcher.UploadStorageFile(id, testfile, "/")
				Expect(err).ToNot(HaveOccurred())
				err = fetcher.UploadStorageFile(id, testfile, "/foo")
				Expect(err).ToNot(HaveOccurred())

				list, err := fetcher.StorageFileList(id)
				Expect(len(list)).To(Equal(2))
				Expect(list[0]).To(Equal("/foo/test"))
				Expect(list[1]).To(Equal("/test"))

				dir2, err := ioutil.TempDir("", "client_test2")
				defer os.RemoveAll(dir2) // clean up

				err = fetcher.DownloadArtefactsFromStorage(id, dir2)
				Expect(err).ToNot(HaveOccurred())

				_, err = os.Stat(filepath.Join(dir2, "test"))
				Expect(err).ToNot(HaveOccurred())

				sum2, err := helpers.FileSum(filepath.Join(dir2, "test"))
				Expect(err).ToNot(HaveOccurred())
				Expect(sum2 == sum).To(Equal(true))

				ev, err = fetcher.StorageDelete(id)
				ExpectSuccessfulResponse(ev, err)

			})
		})

		Context("Token", func() {
			It("Can create and update them", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				ev, err := fetcher.TokenCreate()
				id := ev.ID
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.TokenDelete(id)
				ExpectSuccessfulResponse(ev, err)
			})
		})

		Context("User", func() {
			It("Can create and update them", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				ev, err := fetcher.UserCreate(map[string]interface{}{
					"name":     "test4",
					"username": "foo",
					"password": "barbarbaz",
				})
				ExpectSuccessfulResponse(ev, err)
				ExpectSuccessfulResponse(fetcher.UserRemove(ev.ID))
			})
		})

		Context("WebHooks", func() {
			It("Can create and update them", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				ev, err := fetcher.WebHookCreate("github")
				ExpectSuccessfulResponse(ev, err)
				id := ev.ID
				ev, err = fetcher.WebHookTaskUpdate(id, helpers.FixtureTaskData)
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.WebHookEdit(map[string]interface{}{"id": id, "value": "master", "key": "filter"})
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.WebHookDelete(id)
				ExpectSuccessfulResponse(ev, err)
			})
		})

		Context("Plan", func() {
			It("Can create and delete them", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				ev, err := fetcher.PlanCreate(helpers.FixtureTaskData)
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.PlanDelete(ev.ID)
				ExpectSuccessfulResponse(ev, err)
			})
		})

		Context("Pipeline", func() {
			It("Can create and delete them", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])
				pipeline := &tasks.Pipeline{
					Name: "Test",
					Tasks: map[string]tasks.Task{
						"test": tasks.Task{
							Name: "test",
						},
					},
				}
				ev, err := fetcher.PipelineCreate(pipeline.ToMap(false))
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.PipelineDelete(ev.ID)
				ExpectSuccessfulResponse(ev, err)
			})
		})

		Context("Task", func() {
			It("Can do start/stop/report result", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				ev, err := fetcher.SetupTask()
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.StopTask(helpers.Tasks[0])
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.SetTaskResult("error")
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.SetTaskStatus("done")
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.CloneTask(helpers.Tasks[0])
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.CreateTask(helpers.FixtureTaskData)
				ExpectSuccessfulResponse(ev, err)
				newtask := ev.ID

				ev, err = fetcher.SetTaskField("output", "test")
				ExpectSuccessfulResponse(ev, err)

				ev, err = fetcher.TaskDelete(newtask)
				ExpectSuccessfulResponse(ev, err)
			})
		})

		Context("Secrets", func() {
			It("Can create and update them", func() {
				fetcher, err := NewFakeClient()
				Expect(err).ToNot(HaveOccurred())
				fetcher.Doc(helpers.Tasks[0])

				ev, err := fetcher.SecretCreate("test")
				ExpectSuccessfulResponse(ev, err)
				id := ev.ID

				ev, err = fetcher.SecretEdit(map[string]interface{}{"id": id, "key": "secret", "value": "ok"})
				ExpectSuccessfulResponse(ev, err)

				var s secret.Secret
				req := schema.Request{
					Route:   v1.Schema.GetSecretRoute("show"),
					Target:  &s,
					Options: map[string]interface{}{"id": id},
				}
				err = fetcher.Handle(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.OwnerId).To(Equal(helpers.UserID))
				Expect(s.ID).To(Equal(id))
				Expect(s.Secret).To(Equal("ok"))

				req = schema.Request{
					Route:   v1.Schema.GetSecretRoute("show_by_name"),
					Target:  &s,
					Options: map[string]interface{}{"name": "test"},
				}
				err = fetcher.Handle(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.Secret).To(Equal("ok"))
				Expect(s.OwnerId).To(Equal(helpers.UserID))

				ev, err = fetcher.SecretDelete(id)
				ExpectSuccessfulResponse(ev, err)
			})
		})
	})
})

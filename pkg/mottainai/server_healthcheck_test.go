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

package mottainai_test

import (
	"io/ioutil"
	"os"
	"time"

	node "github.com/MottainaiCI/mottainai-server/pkg/nodes"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	helpers "github.com/MottainaiCI/mottainai-server/tests/helpers"
)

var _ = Describe("Mottainai Healthcheck server", func() {

	dir, _ := ioutil.TempDir("", "healtcheck_test")
	defer os.RemoveAll(dir) // clean up
	helpers.InitConfig(dir)
	m := Classic(helpers.Config)
	db := helpers.InitDB(helpers.Config)

	Describe("Healtchech Server", func() {
		Context("When there is only one that reached deadline", func() {
			It("Aborts", func() {
				dat := make(map[string]interface{})

				var flagsName []string = []string{
					"script", "storage", "source", "directory", "task", "image",
					"namespace", "storage_path", "artefact_path", "tag_namespace",
					"prune", "queue", "cache_image",
				}

				for _, n := range flagsName {
					dat[n] = "test"
				}
				start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Format("20060102150405")

				dat["last_update_time"] = start
				dat["status"] = setting.TASK_STATE_RUNNING

				id, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				t, err := db.Driver.GetTask(helpers.Config, id)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_RUNNING))
				Expect(t.Queue).Should(Equal("test"))

				err = m.CheckTasksDeadline(db, helpers.Config)
				Expect(err).ToNot(HaveOccurred())

				t, err = db.Driver.GetTask(helpers.Config, id)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_STOPPED))
				Expect(t.Result).Should(Equal(setting.TASK_RESULT_ERROR))
				Expect(t.Output).Should(ContainSubstring("Task exceeded deadline"))

			})
		})

		Context("When there is only one that reached deadline, but others running are fine", func() {
			It("Aborts only the one that reached the deadline", func() {
				dat := make(map[string]interface{})

				var flagsName []string = []string{
					"script", "storage", "source", "directory", "task", "image",
					"namespace", "storage_path", "artefact_path", "tag_namespace",
					"prune", "queue", "cache_image",
				}

				for _, n := range flagsName {
					dat[n] = "test"
				}
				start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Format("20060102150405")

				dat["last_update_time"] = start
				dat["status"] = setting.TASK_STATE_RUNNING

				id, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["last_update_time"] = time.Now().Format("20060102150405")
				intact, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())
				intact2, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())
				intact3, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				t, err := db.Driver.GetTask(helpers.Config, id)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_RUNNING))
				Expect(t.Queue).Should(Equal("test"))

				err = m.CheckTasksDeadline(db, helpers.Config)
				Expect(err).ToNot(HaveOccurred())

				t, err = db.Driver.GetTask(helpers.Config, id)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_STOPPED))
				Expect(t.Result).Should(Equal(setting.TASK_RESULT_ERROR))
				Expect(t.Output).Should(ContainSubstring("Task exceeded deadline"))

				// The others should be intact then

				ti, err := db.Driver.GetTask(helpers.Config, intact)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_RUNNING))

				ti, err = db.Driver.GetTask(helpers.Config, intact2)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_RUNNING))

				ti, err = db.Driver.GetTask(helpers.Config, intact3)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_RUNNING))
			})
		})

		Context("When there is only one that reached deadline, but there are tasks in all states", func() {
			It("Aborts only the two that reached the deadline", func() {
				dat := make(map[string]interface{})

				var flagsName []string = []string{
					"script", "storage", "source", "directory", "task", "image",
					"namespace", "storage_path", "artefact_path", "tag_namespace",
					"prune", "queue", "cache_image",
				}

				for _, n := range flagsName {
					dat[n] = "test"
				}
				start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Format("20060102150405")

				dat["last_update_time"] = start
				dat["status"] = setting.TASK_STATE_RUNNING

				id, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())
				id2, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["last_update_time"] = time.Now().Format("20060102150405")
				intact, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())
				dat["status"] = setting.TASK_STATE_WAIT

				intact2, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["status"] = setting.TASK_STATE_DONE
				intact3, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["status"] = setting.TASK_STATE_ASK_STOP
				intact4, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["status"] = setting.TASK_STATE_SETUP
				intact5, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				t, err := db.Driver.GetTask(helpers.Config, id)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_RUNNING))
				Expect(t.Queue).Should(Equal("test"))

				m.CheckTasksDeadline(db, helpers.Config)

				t, err = db.Driver.GetTask(helpers.Config, id)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_STOPPED))
				Expect(t.Result).Should(Equal(setting.TASK_RESULT_ERROR))
				Expect(t.Output).Should(ContainSubstring("Task exceeded deadline"))

				t, err = db.Driver.GetTask(helpers.Config, id2)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_STOPPED))
				Expect(t.Result).Should(Equal(setting.TASK_RESULT_ERROR))
				Expect(t.Output).Should(ContainSubstring("Task exceeded deadline"))

				// The others should be intact then

				ti, err := db.Driver.GetTask(helpers.Config, intact)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_RUNNING))

				ti, err = db.Driver.GetTask(helpers.Config, intact2)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_WAIT))

				ti, err = db.Driver.GetTask(helpers.Config, intact3)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_DONE))

				ti, err = db.Driver.GetTask(helpers.Config, intact4)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_ASK_STOP))

				ti, err = db.Driver.GetTask(helpers.Config, intact5)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_SETUP))
			})
		})

		Context("When there is one node that reached deadline, but there are others ok", func() {
			It("Aborts only the two that reached the deadline", func() {
				dat := make(map[string]interface{})
				past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Format("20060102150405")
				now := time.Now().Format("20060102150405")

				deadnode, err := db.Driver.InsertNode(&node.Node{LastReport: past})
				Expect(err).ToNot(HaveOccurred())

				deadnode2, err := db.Driver.InsertNode(&node.Node{LastReport: past})
				Expect(err).ToNot(HaveOccurred())

				activenode, err := db.Driver.InsertNode(&node.Node{LastReport: now})
				Expect(err).ToNot(HaveOccurred())

				activenode2, err := db.Driver.InsertNode(&node.Node{LastReport: now})
				Expect(err).ToNot(HaveOccurred())

				var flagsName []string = []string{
					"script", "storage", "source", "directory", "task", "image",
					"namespace", "storage_path", "artefact_path", "tag_namespace",
					"prune", "queue", "cache_image",
				}

				for _, n := range flagsName {
					dat[n] = "test"
				}
				// Lets fake what we should have in reality: update maybe was there for few hours, but we didn't heard from the node since long time
				///
				dat["last_update_time"] = now
				dat["status"] = setting.TASK_STATE_RUNNING
				dat["node_id"] = deadnode

				id, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["node_id"] = deadnode2
				id2, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["node_id"] = activenode

				dat["last_update_time"] = time.Now().Format("20060102150405")
				intact, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())
				dat["status"] = setting.TASK_STATE_WAIT
				dat["node_id"] = ""
				intact2, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["status"] = setting.TASK_STATE_DONE
				dat["node_id"] = activenode2

				intact3, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["status"] = setting.TASK_STATE_ASK_STOP
				intact4, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				dat["status"] = setting.TASK_STATE_SETUP
				intact5, err := db.Driver.CreateTask(dat)
				Expect(err).ToNot(HaveOccurred())

				t, err := db.Driver.GetTask(helpers.Config, id)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_RUNNING))
				Expect(t.Queue).Should(Equal("test"))

				err = m.CheckNodesDeadline(db, helpers.Config)
				Expect(err).ToNot(HaveOccurred())

				t, err = db.Driver.GetTask(helpers.Config, id)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_STOPPED))
				Expect(t.Result).Should(Equal(setting.TASK_RESULT_ERROR))
				Expect(t.Output).Should(ContainSubstring("Didn't heard from the node since"))

				t, err = db.Driver.GetTask(helpers.Config, id2)
				Expect(err).ToNot(HaveOccurred())

				Expect(t.Status).Should(Equal(setting.TASK_STATE_STOPPED))
				Expect(t.Result).Should(Equal(setting.TASK_RESULT_ERROR))
				Expect(t.Output).Should(ContainSubstring("Didn't heard from the node since"))

				// The others should be intact then

				ti, err := db.Driver.GetTask(helpers.Config, intact)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_RUNNING))

				ti, err = db.Driver.GetTask(helpers.Config, intact2)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_WAIT))

				ti, err = db.Driver.GetTask(helpers.Config, intact3)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_DONE))

				ti, err = db.Driver.GetTask(helpers.Config, intact4)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_ASK_STOP))

				ti, err = db.Driver.GetTask(helpers.Config, intact5)
				Expect(err).ToNot(HaveOccurred())
				Expect(ti.Status).Should(Equal(setting.TASK_STATE_SETUP))
			})
		})

	})
})

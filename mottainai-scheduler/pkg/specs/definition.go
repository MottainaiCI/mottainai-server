/*
Copyright (C) 2021 Daniele Rondina <geaaru@sabayonlinux.org>

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

package specs

import (
	"github.com/MottainaiCI/mottainai-server/pkg/queues"
)

type TaskScheduler interface {
	Setup() error
	RetrieveDefaultQueue() error
	RetrieveNodes() error
	GetQueues() ([]queues.Queue, error)
	GetTasks2Inject() (map[string]map[string][]string, error)
	AnalyzePipeline(string, queues.Queue, []queues.Queue) error

	Schedule() error
}

type NodeSlots struct {
	Key           string
	AvailableSlot int
}

type NodeSlotsList []NodeSlots

func (p NodeSlotsList) Len() int           { return len(p) }
func (p NodeSlotsList) Less(i, j int) bool { return p[i].AvailableSlot < p[j].AvailableSlot }
func (p NodeSlotsList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

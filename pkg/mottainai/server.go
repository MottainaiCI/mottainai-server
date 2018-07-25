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
	"strconv"
	"time"

	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	machinery "github.com/RichardKnop/machinery/v1"
	results "github.com/RichardKnop/machinery/v1/backends/result"
	machinerytask "github.com/RichardKnop/machinery/v1/tasks"
)

type Broker struct {
	Queue  string
	Server *machinery.Server
}

type BrokerSendOptions struct {
	Delayed  string
	TaskName string
	TaskID   int
}

type MottainaiServer struct {
	Servers map[string]*Broker
}

func NewServer() *MottainaiServer { return &MottainaiServer{Servers: make(map[string]*Broker)} }
func NewBroker() *Broker          { return &Broker{} }

func (s *MottainaiServer) Add(queue string) *Broker {
	broker := NewBroker()
	broker.Queue = queue
	if conn, err := NewMachineryServer(queue); err != nil {
		panic(err)
	} else {
		broker.Server = conn
	}
	th := agenttasks.DefaultTaskHandler()
	th.RegisterTasks(broker.Server)
	s.Servers[queue] = broker
	return broker
}

func (s *MottainaiServer) Get(queue string) *Broker {
	if broker, ok := s.Servers[queue]; ok {
		return broker
	} else {
		return s.Add(queue)
	}
}
func (b *Broker) NewWorker(ID string, parallel int) *machinery.Worker {
	return b.Server.NewWorker(ID, parallel)
}

func (b *Broker) SendTask(opts *BrokerSendOptions) (*results.AsyncResult, error) {
	taskname := opts.TaskName
	taskid := opts.TaskID
	onErr := make([]*machinerytask.Signature, 0)

	onErr = append(onErr, &machinerytask.Signature{
		Name: "error",
		Args: []machinerytask.Arg{
			{
				Type:  "string",
				Value: strconv.Itoa(taskid),
			},
		},
	})

	onSuccess := make([]*machinerytask.Signature, 0)

	onSuccess = append(onSuccess, &machinerytask.Signature{
		Name: "success",
		Args: []machinerytask.Arg{
			{
				Type:  "string",
				Value: strconv.Itoa(taskid),
			},
		},
	})

	signature := &machinerytask.Signature{
		Name:       taskname,
		RetryCount: 0,
		Args: []machinerytask.Arg{
			{
				Type:  "string",
				Value: strconv.Itoa(taskid),
			},
		},
		OnError:   onErr,
		OnSuccess: onSuccess,
	}
	if len(opts.Delayed) > 0 {
		if secs, err := strconv.Atoi(opts.Delayed); err != nil {
			t := time.Now().UTC().Add(time.Duration(secs) * time.Second)
			signature.ETA = &t
		}
	}

	return b.Server.SendTask(signature)

}

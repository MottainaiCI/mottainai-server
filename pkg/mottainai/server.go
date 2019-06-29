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

	config "github.com/RichardKnop/machinery/v1/config"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	taskmanager "github.com/MottainaiCI/mottainai-server/pkg/tasks/manager"
	machinery "github.com/RichardKnop/machinery/v1"
	results "github.com/RichardKnop/machinery/v1/backends/result"
	machinerytask "github.com/RichardKnop/machinery/v1/tasks"
)

type Broker struct {
	Queue  string
	Server *machinery.Server
}

type BrokerSendOptions struct {
	Delayed           string
	Type              string
	TaskID            string
	Group, ChordGroup map[string]string
	Retry             int
	Concurrency       string
}

type MottainaiServer struct {
	Servers map[string]*Broker
}

func NewServer() *MottainaiServer { return &MottainaiServer{Servers: make(map[string]*Broker)} }
func NewBroker() *Broker          { return &Broker{} }

func NewMachineryServer(queue string, settings *setting.Config) (*machinery.Server, error) {
	var cnf = &config.Config{
		Broker:          settings.GetBroker().Broker,
		DefaultQueue:    queue,
		ResultBackend:   settings.GetBroker().BrokerResultBackend,
		ResultsExpireIn: settings.GetBroker().ResultsExpireIn,
		NoUnixSignals:   !settings.GetBroker().HandleSignal,
	}
	switch broker := settings.GetBroker().Type; broker {
	case "amqp":
		cnf.AMQP = &config.AMQPConfig{
			Exchange:     settings.GetBroker().BrokerExchange,
			ExchangeType: settings.GetBroker().BrokerExchangeType,
			BindingKey:   queue + "_key",
			//BindingKey:   settings.BrokerBindingKey,
		}
	case "redis":
		cnf.Redis = &config.RedisConfig{
			MaxIdle:                settings.GetBroker().MaxIdle,
			MaxActive:              settings.GetBroker().MaxActive,
			IdleTimeout:            settings.GetBroker().IdleTimeout,
			Wait:                   settings.GetBroker().Wait,
			ReadTimeout:            settings.GetBroker().ReadTimeout,
			WriteTimeout:           settings.GetBroker().WriteTimeout,
			ConnectTimeout:         settings.GetBroker().ConnectTimeout,
			DelayedTasksPollPeriod: settings.GetBroker().DelayedTasksPollPeriod,
		}
	case "dynamodb":
		cnf.DynamoDB = &config.DynamoDBConfig{
			TaskStatesTable: settings.GetBroker().TaskStatesTable,
			GroupMetasTable: settings.GetBroker().GroupMetasTable,
		}
	}
	return machinery.NewServer(cnf)
}

func (s *MottainaiServer) Add(queue string, config *setting.Config) *Broker {
	broker := NewBroker()
	broker.Queue = queue
	if conn, err := NewMachineryServer(queue, config); err != nil {
		panic(err)
	} else {
		broker.Server = conn
	}
	th := taskmanager.DefaultTaskHandler(config)
	th.RegisterTasks(broker.Server)
	s.Servers[queue] = broker
	return broker
}

func (s *MottainaiServer) Get(queue string, config *setting.Config) *Broker {
	if broker, ok := s.Servers[queue]; ok {
		return broker
	} else {
		return s.Add(queue, config)
	}
}
func (b *Broker) NewWorker(ID string, parallel int) *machinery.Worker {
	return b.Server.NewWorker(ID, parallel)
}
func (b *Broker) SendChain(opts *BrokerSendOptions) (*results.ChainAsyncResult, error) {

	group := make([]*machinerytask.Signature, 0)

	for i, task_type := range opts.Group {
		onErr := make([]*machinerytask.Signature, 0)

		onErr = append(onErr, &machinerytask.Signature{
			Name: "error",
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
		})

		onSuccess := make([]*machinerytask.Signature, 0)

		onSuccess = append(onSuccess, &machinerytask.Signature{
			Name: "success",
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
		})

		signature := &machinerytask.Signature{
			Name:       task_type,
			RetryCount: opts.Retry,
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
			OnError:   onErr,
			OnSuccess: onSuccess,
		}

		group = append(group, signature)
	}

	g, _ := machinerytask.NewChain(group...)
	return b.Server.SendChain(g)
}

func (b *Broker) SendGroup(opts *BrokerSendOptions) ([]*results.AsyncResult, error) {

	group := make([]*machinerytask.Signature, 0)

	for i, task_type := range opts.Group {
		onErr := make([]*machinerytask.Signature, 0)

		onErr = append(onErr, &machinerytask.Signature{
			Name: "error",
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
		})

		onSuccess := make([]*machinerytask.Signature, 0)

		onSuccess = append(onSuccess, &machinerytask.Signature{
			Name: "success",
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
		})

		signature := &machinerytask.Signature{
			Name:       task_type,
			RetryCount: opts.Retry,
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
			OnError:   onErr,
			OnSuccess: onSuccess,
		}

		group = append(group, signature)
	}
	ci, _ := strconv.Atoi(opts.Concurrency)
	g, _ := machinerytask.NewGroup(group...)
	return b.Server.SendGroup(g, ci)
}

func (b *Broker) SendChord(opts *BrokerSendOptions) (*results.ChordAsyncResult, error) {

	group := make([]*machinerytask.Signature, 0)
	chord := make([]*machinerytask.Signature, 0)

	for i, task_type := range opts.Group {
		onErr := make([]*machinerytask.Signature, 0)

		onErr = append(onErr, &machinerytask.Signature{
			Name: "error",
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
		})

		onSuccess := make([]*machinerytask.Signature, 0)

		onSuccess = append(onSuccess, &machinerytask.Signature{
			Name: "success",
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
		})

		signature := &machinerytask.Signature{
			Name:       task_type,
			RetryCount: opts.Retry,
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
			OnError:   onErr,
			OnSuccess: onSuccess,
		}

		group = append(group, signature)
	}

	for i, task_type := range opts.ChordGroup {
		onErr := make([]*machinerytask.Signature, 0)

		onErr = append(onErr, &machinerytask.Signature{
			Name: "error",
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
		})

		onSuccess := make([]*machinerytask.Signature, 0)

		onSuccess = append(onSuccess, &machinerytask.Signature{
			Name: "success",
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
		})

		signature := &machinerytask.Signature{
			Name:       task_type,
			RetryCount: opts.Retry,
			Args: []machinerytask.Arg{
				{
					Type:  "string",
					Value: i,
				},
			},
			OnError:   onErr,
			OnSuccess: onSuccess,
		}

		chord = append(chord, signature)
	}

	g, _ := machinerytask.NewGroup(group...)
	cc, _ := machinerytask.NewChord(g, chord[0]) // Only one is supported..
	ci, _ := strconv.Atoi(opts.Concurrency)

	return b.Server.SendChord(cc, ci)
}

func (b *Broker) SendTask(opts *BrokerSendOptions) (*results.AsyncResult, error) {
	taskname := opts.Type
	taskid := opts.TaskID
	onErr := make([]*machinerytask.Signature, 0)

	onErr = append(onErr, &machinerytask.Signature{
		Name: "error",
		Args: []machinerytask.Arg{
			{
				Type:  "string",
				Value: taskid,
			},
		},
	})

	onSuccess := make([]*machinerytask.Signature, 0)

	onSuccess = append(onSuccess, &machinerytask.Signature{
		Name: "success",
		Args: []machinerytask.Arg{
			{
				Type:  "string",
				Value: taskid,
			},
		},
	})

	signature := &machinerytask.Signature{
		Name:       taskname,
		RetryCount: opts.Retry,
		Args: []machinerytask.Arg{
			{
				Type:  "string",
				Value: taskid,
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

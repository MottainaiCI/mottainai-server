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
	"reflect"
	"testing"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

func TestNew(t *testing.T) {
	server := NewServer()
	broker := NewBroker()

	if reflect.TypeOf(server).String() != "*mottainai.MottainaiServer" {
		t.Errorf("%v returned", server)
	}

	if reflect.TypeOf(broker).String() != "*mottainai.Broker" {
		t.Errorf("%v returned", broker)
	}
}

func TestAdd(t *testing.T) {
	server := NewServer()
	setting.GenDefault()

	server.Add("test")

	if reflect.TypeOf(server.Servers["test"]).String() != "*mottainai.Broker" {
		t.Errorf("N%v returned", server.Servers["test"])
	}

	if len(server.Servers) != 1 {
		t.Errorf("One server by default")
	}
}

func TestGet(t *testing.T) {
	server := NewServer()
	setting.GenDefault()

	if len(server.Servers) != 0 {
		t.Errorf("0 server by default")
	}
	server.Add("test")

	if reflect.TypeOf(server.Servers["test"]).String() != "*mottainai.Broker" {
		t.Errorf("N%v returned", server.Servers["test"])
	}

	if len(server.Servers) != 1 {
		t.Errorf("One server by default")
	}
}

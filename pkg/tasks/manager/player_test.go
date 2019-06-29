/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package agenttasks

import (
	"errors"
	"testing"
)

type TestExecutor struct {
	Failed          bool
	ExitStatusFired bool
	SuccessCode     int
	Cleanup         bool
}

func (t *TestExecutor) Play(s string) (int, error) {
	if s == "foo" {
		return 0, nil
	}
	return 1, errors.New("Foo was not entered!")
}

func (t *TestExecutor) Setup(s string) error {
	if s == "foo" {
		return nil
	}
	return errors.New("Setup failed!")
}

func (t *TestExecutor) Clean() error {
	t.Cleanup = true

	return errors.New("Cleanup Fails")
}

func (t *TestExecutor) Fail(s string) {
	t.Failed = true
}
func (t *TestExecutor) Report(v ...interface{}) {

}
func (t *TestExecutor) Success(e int) {
	t.SuccessCode = e
}
func (t *TestExecutor) ExitStatus(i int) {
	t.ExitStatusFired = true
}

func TestPlayer(t *testing.T) {

	te := &TestExecutor{}

	player := &Player{TaskID: "foo"}

	i, err := player.Start(te)
	if err != nil {
		t.Error(err)
	}
	if !te.Cleanup {
		t.Error("Cleanup phase didn't fired")
	}
	if te.SuccessCode != 0 {
		t.Error("Player should have exited successfully")
	}

	player = &Player{TaskID: "bad"}

	i, err = player.Start(te)
	if !te.Cleanup {
		t.Error("Cleanup phase didn't fired")
	}
	if err.Error() != "Setup phase error: Setup failed!" {
		t.Error(err.Error())
	}

	if i != 1 {
		t.Error("Player didn't failed")
	}

}

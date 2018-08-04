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
	"strconv"
)

type Executor interface {
	Play(string) (int, error)
	Setup(string) error
	Clean() error
	Fail(string)
	Success(int)
	ExitStatus(int)
}

type Player struct{ TaskID string }

func NewPlayer(taskid string) *Player {
	return &Player{TaskID: taskid}
}

func (p *Player) EarlyFail(e Executor, TaskID, reason string) {
	err := e.Setup(p.TaskID)
	if err != nil {
		e.Fail("Setup phase error: " + err.Error())
	}
	e.Fail(reason)
}

func (p *Player) Start(e Executor) (int, error) {
	defer e.Clean()
	err := e.Setup(p.TaskID)
	if err != nil {
		e.Fail("Setup phase error: " + err.Error())
		return 1, errors.New("Setup phase error: " + err.Error())
	}

	res, err := e.Play(p.TaskID)

	if err != nil {
		if err.Error() == ABORT_EXECUTION_ERROR {
			return 0, nil
		}
		errmsg := "Play phase error (Exit with: " + strconv.Itoa(res) + ") : " + err.Error()
		e.Fail(errmsg)
		return 1, errors.New(errmsg)
	} else {
		e.Success(res)
	}

	e.ExitStatus(res)

	return res, err
}

// +build lxd

/*

Copyright (C) 2017-2021  Daniele Rondina <geaaru@sabayonlinux.org>
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
	"io"

	lxd_compose "github.com/MottainaiCI/lxd-compose/pkg/executor"
)

// We don't use host commands
func (l *LxdExecutor) GetHostWriterStdout() io.WriteCloser  { return nil }
func (l *LxdExecutor) GetHostWriterStderr() io.WriteCloser  { return nil }
func (l *LxdExecutor) SetHostWriterStdout(w io.WriteCloser) {}
func (l *LxdExecutor) SetHostWriterStderr(w io.WriteCloser) {}

func (l *LxdExecutor) GetLxdWriterStdout() io.WriteCloser  { return l }
func (l *LxdExecutor) GetLxdWriterStderr() io.WriteCloser  { return l }
func (l *LxdExecutor) SetLxdWriterStdout(w io.WriteCloser) {}
func (l *LxdExecutor) SetLxdWriterStderr(w io.WriteCloser) {}

func (l *LxdExecutor) DebugLog(color bool, args ...interface{}) {
	l.Report(args...)
}

func (l *LxdExecutor) InfoLog(color bool, args ...interface{}) {
	l.Report(args...)
}

func (l *LxdExecutor) WarnLog(color bool, args ...interface{}) {
	l.Report(args...)
}

func (l *LxdExecutor) ErrorLog(color bool, args ...interface{}) {
	l.Report(args...)
}

func (l *LxdExecutor) Emits(eType lxd_compose.LxdCExecutorEvent, data map[string]interface{}) {}

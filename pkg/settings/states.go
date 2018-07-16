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

package setting

const TASK_STATE_RUNNING = "running"
const TASK_STATE_DONE = "done"
const TASK_STATE_STOPPED = "stopped"
const TASK_STATE_ASK_STOP = "stop"
const TASK_STATE_WAIT = "waiting"

const TASK_RESULT_FAILED = "failed"
const TASK_RESULT_ERROR = "error"
const TASK_RESULT_SUCCESS = "success"
const TASK_RESULT_UNKNOWN = "none"

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

package client

import (
	"fmt"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

func (f *Fetcher) SetTaskField(field, value string) ([]byte, error) {
	return f.GetOptions("/api/tasks/updatefield", map[string]string{
		"id":     f.docID,
		"apikey": setting.Configuration.AgentKey,
		"field":  field,
		"value":  value,
	})
}

func (f *Fetcher) SetTaskStatus(status string) ([]byte, error) {
	return f.GetOptions("/api/tasks/update", map[string]string{
		"id":     f.docID,
		"apikey": setting.Configuration.AgentKey,
		"status": status,
	})
}

func (f *Fetcher) AbortTask() {
	f.SetTaskResult("")
	f.SetTaskStatus("stopped")
}

func (f *Fetcher) FailTask(e string) {
	f.SetTaskStatus("failure")
	f.AppendTaskOutput(e)
}

func (f *Fetcher) SetupTask() {
	f.SetTaskStatus("setup")
}

func (f *Fetcher) RunTask() {
	f.SetTaskStatus("running")
}

func (f *Fetcher) GetTask() ([]byte, error) {
	doc, err := f.GetOptions("/api/tasks/"+f.docID, map[string]string{})
	if err != nil {
		return []byte{}, err
	}
	return doc, err
}

func (f *Fetcher) AllTasks() ([]byte, error) {
	doc, err := f.GetOptions("/api/tasks", map[string]string{})
	if err != nil {
		return []byte{}, err
	}
	return doc, err
}

func (f *Fetcher) SetTaskResult(result string) ([]byte, error) {
	return f.GetOptions("/api/tasks/update", map[string]string{
		"id":     f.docID,
		"apikey": setting.Configuration.AgentKey,
		"result": result,
	})
}

func (f *Fetcher) SetTaskOutput(output string) ([]byte, error) {
	return f.GetOptions("/api/tasks/update", map[string]string{
		"id":     f.docID,
		"apikey": setting.Configuration.AgentKey,
		"output": output,
	})
}

func (f *Fetcher) AppendTaskOutput(output string) ([]byte, error) {
	if f.docID == "" {
		fmt.Println(output)
	}
	return f.GetOptions("/api/tasks/append", map[string]string{
		"id":     f.docID,
		"apikey": setting.Configuration.AgentKey,
		"output": output,
	})
}

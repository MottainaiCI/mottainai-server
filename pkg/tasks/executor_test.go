// +build !all

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
	"io"
	"os"
	"testing"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type FakeClient struct {
	Taskfield1 string
	Taskfield2 string
}

func (f *FakeClient) RunTask()                                            {}
func (f *FakeClient) SetupTask()                                          {}
func (f *FakeClient) Doc(string)                                          {}
func (f *FakeClient) AppendTaskOutput(string) ([]byte, error)             { return []byte{}, nil }
func (f *FakeClient) GetTask() ([]byte, error)                            { return []byte{}, nil }
func (f *FakeClient) AbortTask()                                          {}
func (f *FakeClient) DownloadArtefactsFromTask(string, string) error      { return nil }
func (f *FakeClient) DownloadArtefactsFromNamespace(string, string) error { return nil }
func (f *FakeClient) DownloadArtefactsFromStorage(string, string) error   { return nil }
func (f *FakeClient) UploadFile(string, string) error                     { return nil }
func (f *FakeClient) FailTask(string)                                     {}
func (f *FakeClient) ErrorTask()                                          {}
func (f *FakeClient) RegisterNode(string, string) ([]byte, error)         { return []byte{}, nil }
func (f *FakeClient) FinishTask()                                         {}
func (f *FakeClient) SuccessTask()                                        {}
func (f *FakeClient) StreamOutput(io.Reader)                              {}
func (f *FakeClient) SetTaskField(a string, b string) ([]byte, error) {
	f.Taskfield1 = a
	f.Taskfield2 = b
	return []byte{}, nil
}

func TestTaskExecutor(t *testing.T) {
	t.Parallel()
	config := setting.NewConfig(nil)
	// Set env variable
	config.Viper.SetEnvPrefix(setting.MOTTAINAI_ENV_PREFIX)
	config.Viper.AutomaticEnv()
	config.Viper.SetTypeByDefaultValue(true)
	config.Unmarshal()

	ctx := NewExecutorContext()

	dir := os.TempDir()
	defer os.RemoveAll(dir)

	config.AppName = "test"
	//TODO: Slim context :(
	ctx.ArtefactDir = dir
	ctx.BuildDir = dir
	ctx.StorageDir = dir
	ctx.NamespaceDir = dir
	ctx.BuildDir = dir
	ctx.SourceDir = dir
	ctx.RootTaskDir = dir
	ctx.RealRootDir = dir
	ctx.DocID = "foo"
	f := &FakeClient{}
	e := &TaskExecutor{Context: ctx, Config: config}
	e.MottainaiClient = f
	e.ExitStatus(20)
	// TODO : Replace hardcoded exit_status with reflect lookup on map tag
	if f.Taskfield1 != "exit_status" {
		t.Error("Failed first field encode")
	}

	if f.Taskfield2 != "20" {
		t.Error("Failed second field encode")
	}

}

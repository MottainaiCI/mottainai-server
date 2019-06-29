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
	"fmt"
	"testing"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

func TestPipeline(t *testing.T) {

	config := setting.NewConfig(nil)
	// Set env variable
	config.Viper.SetEnvPrefix(setting.MOTTAINAI_ENV_PREFIX)
	config.Viper.AutomaticEnv()
	config.Viper.SetTypeByDefaultValue(true)
	config.Unmarshal()

	pipe := &Pipeline{}
	pipe.Group = []string{"test", "test1"}

	m := make(map[string]Task)
	test := Task{Namespace: "boh"}
	test1 := Task{Namespace: "bar"}
	m["test"] = test
	m["test1"] = test1

	pipe.Tasks = m

	test_res := pipe.ToMap(true)
	fmt.Println(test_res)

	pipe2 := NewPipelineFromMap(test_res)

	if pipe2.Tasks["test"].Namespace != "boh" {
		t.Error("Invalid namespace for ", pipe2.Tasks["test"])
	}
	if pipe2.Tasks["test1"].Namespace != "bar" {
		t.Error("Invalid namespace for ", pipe2.Tasks["test1"])
	}
}

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
package main

import (
	"os"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	"github.com/MottainaiCI/mottainai-server/cmd"

	"github.com/urfave/cli"
)

func init() {
	setting.AppVer = setting.MOTTAINAI_VERSION
	setting.HTTPAddr = "127.0.0.1"
	setting.HTTPPort = "9090"
	setting.Protocol = "http"
	setting.AppName = "Mottainai"
	setting.AppURL = "http://127.0.0.1:9090"
	setting.SecretKey = "baijoibejoiebgjoi"
	setting.StaticRootPath = "./"
	setting.CustomPath = "./"
}

func main() {
	app := cli.NewApp()
	app.Name = "Mottainai"
	app.Usage = "Task/Job Build Service"
	app.Version = setting.MOTTAINAI_VERSION
	app.Commands = []cli.Command{
		cmd.Web,
		cmd.WebHook,
		cmd.Daemon,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}

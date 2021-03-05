/*

Copyright (C) 2017-2021  Ettore Di Giacinto <mudler@gentoo.org>

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
package common

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	event "github.com/MottainaiCI/mottainai-server/pkg/event"

	cobra "github.com/spf13/cobra"
	v "github.com/spf13/viper"
)

func PrintBuff(buff []byte) {
	data := string(buff)
	data = strings.TrimSpace(data)
	if len(data) > 0 {
		fmt.Println(data)
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func GetHomeDir() string {
	u, _ := user.Current()
	ans := u.HomeDir
	if os.Getenv("HOME") != "" {
		ans = os.Getenv("HOME")
	}
	return ans
}

func PrintResponse(resp event.APIResponse) {
	if len(resp.Error) > 0 {
		fmt.Println("ERROR:")
		fmt.Println(resp.Error)
	}
	if len(resp.Data) > 0 {
		fmt.Println("DATA:")
		fmt.Println(resp.Data)
	}
	if len(resp.Processed) > 0 {
		fmt.Println("Processed: " + resp.Processed)
	}
	if len(resp.Status) > 0 {
		fmt.Println("Status: " + resp.Status)
	}
	if len(resp.ObjType) > 0 {
		fmt.Println("ObjType: " + resp.ObjType)
	}
	if len(resp.Event) > 0 {
		fmt.Println("Event: " + resp.Event)
	}
	if len(resp.ID) > 0 {
		fmt.Println("ID: " + resp.ID)
	}
}

// TODO: pass settings in input.
func BuildCmdArgs(cmd *cobra.Command, msg string) string {
	var ans string = "mottainai-cli "

	if cmd == nil {
		panic("Invalid command")
	}

	if cmd.Flag("master").Changed {
		ans += "--master " + v.GetString("master") + " "
	}
	if v.GetString("profile") != "" {
		ans += "--profile " + v.GetString("profile") + " "
	}

	ans += msg

	return ans
}

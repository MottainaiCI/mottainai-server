/*

Copyright (C) 2019 Ettore Di Giacinto <mudler@gentoo.org>

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

package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	template "github.com/MottainaiCI/lxd-compose/pkg/template"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
)

func newCompileCommand(config *setting.Config) *cobra.Command {
	var values []string

	var cmd = &cobra.Command{
		Use:   "compile <template.tmpl> -s foo=1 -s foo=2 [-o task.yaml] [-l values.yaml]",
		Short: "compile a template to a task",
		//		Args:  cobra.RangeArgs(1, 1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Not enough arguments to compile")
				os.Exit(1)
			}

			va := map[string]interface{}{}

			for _, v := range values {
				item := strings.Split(v, "=")
				if len(item) == 0 {
					fmt.Println("Invalid value: ", item)
					os.Exit(1)
				}
				va[item[0]] = strings.Join(item[1:], "=")
			}

			templ := template.NewTemplate()
			templ.Values = va

			vFile, err := cmd.Flags().GetString("load")
			if err != nil {
				fmt.Println("Error loading values file: ", err.Error())
				os.Exit(1)
			}

			if vFile != "" {
				err := templ.LoadValuesFromFile(vFile)
				if err != nil {
					fmt.Println("Error loading values from file: ", err.Error())
					return
				}
			}

			compiled, err := templ.DrawFromFile(args[0])
			if err != nil {
				fmt.Println("Error compiling template: ", err.Error())
				os.Exit(1)
			}

			oFile, err := cmd.Flags().GetString("output")
			if err != nil || oFile == "" {
				fmt.Println(compiled)
				return
			}

			dir := filepath.Dir(oFile)
			os.MkdirAll(dir, 0740)

			f, err := os.Create(oFile)
			if err != nil {
				fmt.Println("Error creating output file: ", err.Error())
				os.Exit(1)
			}
			defer f.Close()

			bytesConsumed, err := f.WriteString(compiled)
			if err != nil {
				fmt.Println("Error writing to output file: ", err.Error())
				os.Exit(1)
			}

			fmt.Printf("wrote %d bytes\n", bytesConsumed)
			f.Sync()
		},
	}

	cmd.Flags().StringP("load", "l", "", "values file")
	cmd.Flags().StringP("output", "o", "", "output file")
	cmd.Flags().StringArrayVarP(&values, "set", "s", []string{}, "")
	return cmd
}

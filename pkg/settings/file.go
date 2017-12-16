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

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// ReadFromFile reads data from a file
func ReadFromFile(cnfPath string) ([]byte, error) {
	file, err := os.Open(cnfPath)

	// Config file not found
	if err != nil {
		return nil, fmt.Errorf("Open file error: %s", err)
	}

	// Config file found, let's try to read it
	data := make([]byte, 1000)
	count, err := file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("Read from file error: %s", err)
	}

	return data[:count], nil
}

func fromFile(cnfPath string) (*Config, error) {
	var newCnf Config
	newCnf = *Configuration

	data, err := ReadFromFile(cnfPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &newCnf); err != nil {
		return nil, fmt.Errorf("Unmarshal YAML error: %s", err)
	}

	return &newCnf, nil
}

func LoadFromFile(cnfPath string) error {
	cfg, err := fromFile(cnfPath)
	if err == nil {
		Configuration = cfg
	}
	return err
}

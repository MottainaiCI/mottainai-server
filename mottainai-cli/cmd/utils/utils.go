/*

Copyright (C) 2017-2021  Daniele Rondina <geaaru@sabayonlinux.org>

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

package utils

import (
	"errors"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	viper "github.com/spf13/viper"
)

func CreateClient(config *setting.Config) (client.HttpClient, error) {
	var v *viper.Viper = config.Viper

	apiKey := v.GetString("apikey")
	apiUser := v.GetString("apiUser")
	apiPass := v.GetString("apiPass")

	if apiKey == "" && (apiUser == "" || apiPass == "") {
		return nil, errors.New("No token or credentials defined")
	}

	if v.GetString("master") == "" {
		return nil, errors.New("Invalid server URL")
	}

	var ans client.HttpClient
	if apiKey != "" {
		ans = client.NewTokenClient(v.GetString("master"), apiKey, config)
	} else {
		ans = client.NewBasicAuthClient(v.GetString("master"), apiUser, apiPass, config)
	}

	return ans, nil
}

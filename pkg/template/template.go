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

package template

import (
	"fmt"
	"html/template"
	"math"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	macaron "gopkg.in/macaron.v1"

	"github.com/Masterminds/sprig"
	"github.com/microcosm-cc/bluemonday"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/markup"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

func TemplatePreview(c *context.Context, templatename string, config *setting.Config) {
	//c.Data["User"] = models.User{Name: "Unknown"}
	c.Data["AppName"] = config.GetWeb().AppName
	c.Data["AppVer"] = setting.MOTTAINAI_VERSION
	c.Data["AppURL"] = config.GetWeb().AppURL
	c.Data["Code"] = "2014031910370000009fff6782aadb2162b4a997acb69d4400888e0b9274657374"
	//c.Data["ActiveCodeLives"] = config.Service.ActiveCodeLives / 60
	//c.Data["ResetPwdCodeLives"] = config.Service.ResetPwdCodeLives / 60
	//c.Data["CurDbValue"] = ""
	c.HTML(200, templatename)
	//c.HTML(200, (c.Params("*")))
}

func Setup(m *macaron.Macaron) {
	m.Invoke(func(config *setting.Config) {
		funcMap := NewFuncMap(config)
		m.Use(macaron.Renderer(macaron.RenderOptions{
			Directory:         path.Join(config.GetWeb().StaticRootPath, "templates"),
			AppendDirectories: []string{path.Join(config.GetWeb().TemplatePath, "templates")},
			Funcs:             funcMap,
			IndentJSON:        macaron.Env != macaron.PROD,
		}))
	})
}

func plural(count int, e string) (ans string) {
	if (count == 1) || (count == 0) {
		ans = strconv.Itoa(count) + " " + e + " "
	} else {
		ans = strconv.Itoa(count) + " " + e + "s "
	}
	return
}

// TODO: only initialize map once and save to a local variable to reduce copies.
func NewFuncMap(config *setting.Config) []template.FuncMap {
	tf := sprig.HtmlFuncMap()
	//tf := sprig.TxtFuncMap()
	tf["GoVer"] = func() string {
		return strings.Title(runtime.Version())
	}
	tf["UseHTTPS"] = func() bool {
		return strings.HasPrefix(config.GetWeb().AppURL, "https")
	}
	tf["AppName"] = func() string {
		return config.GetWeb().AppName
	}
	tf["AppSubURL"] = func() string {
		return config.GetWeb().AppSubURL
	}
	tf["AppURL"] = func() string {
		return config.GetWeb().AppURL
	}
	tf["AppVer"] = func() string {
		return setting.MOTTAINAI_VERSION
	}
	tf["BuildURI"] = func(pattern string) string {
		return config.GetWeb().BuildURI(pattern)
	}
	tf["LoadTimes"] = func(startTime time.Time) string {
		return fmt.Sprint(time.Since(startTime).Nanoseconds()/1e6) + "ms"
	}
	tf["HumanTimeDiff"] = func(startTime, endTime string) string {
		ans := "-"
		s, e1 := time.Parse(setting.Timeformat, startTime)
		e, e2 := time.Parse(setting.Timeformat, endTime)

		if e1 == nil && e2 == nil {
			diff := int(e.Unix() - s.Unix())
			seconds := diff % (60 * 60 * 24 * 7)
			days := math.Floor(float64(seconds) / 60 / 60 / 24)
			seconds = diff % (60 * 60 * 24)
			hours := math.Floor(float64(seconds) / 60 / 60)
			seconds = diff % (60 * 60)
			minutes := math.Floor(float64(seconds) / 60)
			seconds = diff % 60
			if days > 0 {
				ans = plural(int(days), "day") + plural(int(hours), "hour") + strconv.Itoa(int(minutes)) + "min " + strconv.Itoa(seconds) + "s"
			} else if hours > 0 {
				ans = plural(int(hours), "hour") + strconv.Itoa(int(minutes)) + "min " + strconv.Itoa(seconds) + "s"
			} else if minutes > 0 {
				ans = strconv.Itoa(int(minutes)) + "min " + strconv.Itoa(seconds) + "s"
			} else {
				ans = strconv.Itoa(seconds) + "s"
			}
		}
		return ans
	}
	tf["Safe"] = Safe
	tf["Sanitize"] = bluemonday.UGCPolicy().Sanitize
	tf["Str2html"] = Str2html
	tf["Add"] = func(a, b int) int {
		return a + b
	}
	tf["SubStr"] = func(str string, start, length int) string {
		if len(str) == 0 {
			return ""
		}
		end := start + length
		if length == -1 {
			end = len(str)
		}
		if len(str) < end {
			return str
		}
		return str[start:end]
	}
	tf["Join"] = strings.Join
	tf["Sha1"] = Sha1
	tf["ShortSHA1"] = utils.ShortSHA1
	tf["MD5"] = utils.MD5
	tf["GenAvatar"] = func(name string, size int) string {
		a := NewGetAvataaarsCom()
		return a.GetAvatar(name)
		//return "https://avatars.moe/Default/" + string(utils.MD5(name)) + "/" + strconv.Itoa(size) + ".jpg"
	}
	return []template.FuncMap{tf}
}

func Safe(raw string) template.HTML {
	return template.HTML(raw)
}

func Str2html(raw string) template.HTML {
	return template.HTML(markup.Sanitize(raw))
}
func Sha1(str string) string {
	return utils.SHA1(str)
}

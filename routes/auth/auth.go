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

package auth

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	com "github.com/Unknwon/com"
	"github.com/go-macaron/binding"
	"github.com/markbates/goth"

	"github.com/markbates/goth/providers/github"
	macaron "gopkg.in/macaron.v1"
)

type SignIn struct {
	UserName    string `binding:"Required;MaxSize(254)"`
	Password    string `binding:"Required;MaxSize(255)"`
	LoginSource int64
	Remember    bool
}

func (f *SignIn) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type Register struct {
	UserName string `binding:"Required;AlphaDashDot;MaxSize(35)"`
	Email    string `binding:"Required;Email;MaxSize(254)"`
	Password string `binding:"Required;MaxSize(255)"`
	Retype   string
}

func (f *Register) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type Form interface {
	binding.Validator
}

// Assign assign form values back to the template data.
func Assign(form interface{}, data map[string]interface{}) {
	typ := reflect.TypeOf(form)
	val := reflect.ValueOf(form)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		fieldName := field.Tag.Get("form")
		// Allow ignored fields in the struct
		if fieldName == "-" {
			continue
		} else if len(fieldName) == 0 {
			fieldName = com.ToSnakeCase(field.Name)
		}

		data[fieldName] = val.Field(i).Interface()
	}
}
func getRuleBody(field reflect.StructField, prefix string) string {
	for _, rule := range strings.Split(field.Tag.Get("binding"), ";") {
		if strings.HasPrefix(rule, prefix) {
			return rule[len(prefix) : len(rule)-1]
		}
	}
	return ""
}

const ERR_ALPHA_DASH_DOT_SLASH = "AlphaDashDotSlashError"

var AlphaDashDotSlashPattern = regexp.MustCompile("[^\\d\\w-_\\./]")

func init() {
	binding.SetNameMapper(com.ToSnakeCase)
	binding.AddRule(&binding.Rule{
		IsMatch: func(rule string) bool {
			return rule == "AlphaDashDotSlash"
		},
		IsValid: func(errs binding.Errors, name string, v interface{}) (bool, binding.Errors) {
			if AlphaDashDotSlashPattern.MatchString(fmt.Sprintf("%v", v)) {
				errs.Add([]string{name}, ERR_ALPHA_DASH_DOT_SLASH, "AlphaDashDotSlash")
				return false, errs
			}
			return true, errs
		},
	})
}

func getSize(field reflect.StructField) string {
	return getRuleBody(field, "Size(")
}

func getMinSize(field reflect.StructField) string {
	return getRuleBody(field, "MinSize(")
}

func getMaxSize(field reflect.StructField) string {
	return getRuleBody(field, "MaxSize(")
}

func getInclude(field reflect.StructField) string {
	return getRuleBody(field, "Include(")
}
func validate(errs binding.Errors, data map[string]interface{}, f Form, l macaron.Locale) binding.Errors {
	if errs.Len() == 0 {
		return errs
	}

	data["HasError"] = true
	Assign(f, data)

	typ := reflect.TypeOf(f)
	val := reflect.ValueOf(f)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		fieldName := field.Tag.Get("form")
		// Allow ignored fields in the struct
		if fieldName == "-" {
			continue
		}

		if errs[0].FieldNames[0] == field.Name {
			data["Err_"+field.Name] = true

			trName := field.Tag.Get("locale")
			if len(trName) == 0 {
				trName = "form." + field.Name
			} else {
				trName = trName
			}

			switch errs[0].Classification {
			case binding.ERR_REQUIRED:
				data["ErrorMsg"] = trName + "form.require_error"
			case binding.ERR_ALPHA_DASH:
				data["ErrorMsg"] = trName + "form.alpha_dash_error"
			case binding.ERR_ALPHA_DASH_DOT:
				data["ErrorMsg"] = trName + "form.alpha_dash_dot_error"
			case ERR_ALPHA_DASH_DOT_SLASH:
				data["ErrorMsg"] = trName + "form.alpha_dash_dot_slash_error"
			case binding.ERR_SIZE:
				data["ErrorMsg"] = trName + "form.size_error" + getSize(field)
			case binding.ERR_MIN_SIZE:
				data["ErrorMsg"] = trName + "form.min_size_error" + getMinSize(field)
			case binding.ERR_MAX_SIZE:
				data["ErrorMsg"] = trName + "form.max_size_error" + getMaxSize(field)
			case binding.ERR_EMAIL:
				data["ErrorMsg"] = trName + "form.email_error"
			case binding.ERR_URL:
				data["ErrorMsg"] = trName + "form.url_error"
			case binding.ERR_INCLUDE:
				data["ErrorMsg"] = trName + "form.include_error" + getInclude(field)
			default:
				data["ErrorMsg"] = "form.unknown_error" + " " + errs[0].Classification
			}
			return errs
		}
	}
	return errs
}

func Setup(m *macaron.Macaron) {
	// ***** START: User *****
	//
	// reqSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: true})
	// ignSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: false})
	// ignSignInAndCsrf := context.Toggle(&context.ToggleOptions{DisableCSRF: true})

	m.Invoke(func(config *setting.Config) {
		reqSignOut := context.Toggle(&context.ToggleOptions{SignOutRequired: true, BaseURL: config.AppSubURL})
		bindIgnErr := binding.BindIgnErr
		reqSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: true, BaseURL: config.AppSubURL})
		reqAdmin := context.Toggle(&context.ToggleOptions{AdminRequired: true, BaseURL: config.AppSubURL})
		reqManager := context.Toggle(&context.ToggleOptions{ManagerRequired: true, BaseURL: config.AppSubURL})

		m.Group("/user", func() {
			m.Group("/login", func() {
				m.Combo("").Get(Login).
					Post(bindIgnErr(SignIn{}), LoginPost)
			})
			m.Get("/sign_up", SignUp)
			m.Post("/sign_up", bindIgnErr(Register{}), SignUpPost)

		}, reqSignOut)

		// TODO: Move from Here
		goth.UseProviders(
			github.New(setting.Configuration.WebHookGitHubToken,
				setting.Configuration.WebHookGitHubSecret,
				setting.Configuration.AppURL+"/auth/github/callback"),
		)

		m.Get("/auth/github/callback", RequiresIntegrationSetting, reqSignIn, GithubAuthCallback)
		m.Get("/logout/github", RequiresIntegrationSetting, reqSignIn, GithubLogout)
		m.Get("/auth/github", RequiresIntegrationSetting, reqSignIn, GithubLogin)

		m.Get("/user/list", reqSignIn, reqManager, ListUsers)
		m.Get("/user/show/:id", reqSignIn, reqManager, Show)

		m.Get("/user/set/admin/:id", reqSignIn, reqAdmin, SetAdmin)
		m.Get("/user/unset/admin/:id", reqSignIn, reqAdmin, UnSetAdmin)
		m.Get("/user/set/manager/:id", reqSignIn, reqAdmin, SetManager)
		m.Get("/user/unset/manager/:id", reqSignIn, reqAdmin, UnSetManager)
		m.Get("/user/delete/:id", reqSignIn, reqAdmin, DeleteUser)

		m.Group("/user", func() {
			m.Get("/logout", SignOut)
		})
	})
}

func WrapF(f http.HandlerFunc) macaron.Handler {
	return func(c *context.Context) {
		f(c.Resp, c.Req.Request)
	}
}

func WrapH(h http.Handler) macaron.Handler {
	return func(c *context.Context) {
		h.ServeHTTP(c.Resp, c.Req.Request)
	}
}

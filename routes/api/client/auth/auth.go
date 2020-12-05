/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
Copyright (C) 2020       Adib Saad <adib.saad@gmail.com>
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
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	"github.com/Unknwon/com"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/captcha"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"
	"reflect"
	"regexp"
	"strings"

	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
)

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

// todo better validation handling and messages
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

type SignIn struct {
	UserName    string `binding:"Required;MaxSize(254)"`
	Password    string `binding:"Required;MaxSize(255)"`
	LoginSource int64
	Remember    bool
}

func (f *SignIn) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type SignUp struct {
	UserName        string `binding:"Required;AlphaDashDot;MaxSize(35)"`
	Email           string `binding:"Required;Email;MaxSize(254)"`
	Password        string `binding:"Required;MaxSize(255)"`
	PasswordConfirm string
	Captcha         string `binding:"Required;MaxSize(10)"`
	CaptchaId       string `binding:"Required;MaxSize(30)" json:"captcha_id"`
}

func (f *SignUp) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type UserResp struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Admin   string `json:"is_admin"`
	Manager string `json:"is_manager"`
}

type ErrorResp struct {
	Error string `json:"error"`
}

func Register(c *context.Context, cpt *captcha.Captcha, f SignUp, db *database.Database) {
	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_SIGNUP_ENABLED)
	if err == nil {
		if uuu.IsDisabled() {
			c.JSON(503, map[string]string{"error": "Signup disabled"})
			return
		}
	}

	if c.HasError() {
		c.JSON(400, map[string]string{"error": c.Flash.ErrorMsg})
		return
	}

	if !cpt.Verify(f.CaptchaId, f.Captcha) {
		c.JSON(400, map[string]string{"error": "Captcha verification failed"})
		return
	}

	if f.Password != f.PasswordConfirm {
		c.JSON(400, map[string]string{"error": "Passwords don't match"})
		return
	}

	// todo: create namespace on registration?
	//check := namespace.Namespace{Name: f.UserName}
	//if check.Exists() {
	//		c.RenderWithErr("Username taken as namespace, pick another one", SIGNUP)
	//		return
	//	}
	u := &user.User{
		Name:     f.UserName,
		Email:    f.Email,
		Password: f.Password,
		// todo: email activation?
		//IsActive: !setting.Service.RegisterEmailConfirm,
	}
	if db.Driver.CountUsers() == 0 {
		u.MakeAdmin() // XXX: ugly, also fix error
	}
	if _, err := db.Driver.InsertAndSaltUser(u); err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	log.Trace("Account created: %s", u.Name)
	c.JSONSuccess(struct{}{})
}

func Login(c *context.Context, f SignIn, db *database.Database) {
	var err error
	var u user.User

	if c.HasError() {
		c.JSON(400, map[string]string{"error": c.Flash.ErrorMsg})
		return
	}

	onlyuser_val, err := db.Driver.GetSettingByKey(
		setting.SYSTEM_SIGNIN_ONLY_USERVALIDATION)
	if err == nil {
		if onlyuser_val.IsEnabled() {
			u, err = db.Driver.GetUserByName(f.UserName)
		} else {
			u, err = db.Driver.SignIn(f.UserName, f.Password)
		}
	} else {
		u, err = db.Driver.SignIn(f.UserName, f.Password)
	}

	if err != nil {
		c.JSON(400, ErrorResp{err.Error()})
		return
	}

	c.Invoke(func(config *setting.Config) {
		if f.Remember {
			days := 86400 * 30
			c.SetCookie("u_name", u.Name, days, config.GetWeb().AppSubURL, "", true, true)
			c.SetSuperSecureCookie(u.Password, "r_name", u.Name, days, config.GetWeb().AppSubURL, "", true, true)
		}

		c.Session.Set("uid", u.ID)
		c.Session.Set("uname", u.Name)

		// Clear whatever CSRF has right now, force to generate a new one
		c.SetCookie("_csrf", "", -1, config.GetWeb().AppSubURL)
	})

	c.JSONSuccess(&u)
}

func Logout(c *context.Context) {
	c.Invoke(func(config *setting.Config) {
		c.Session.Delete("uid")
		c.Session.Delete("uname")
		c.SetCookie("u_name", "", -1, config.GetWeb().AppSubURL)
		c.SetCookie("r_name", "", -1, config.GetWeb().AppSubURL)
		c.SetCookie("_csrf", "", -1, config.GetWeb().AppSubURL)
		c.SubURLRedirect("/")
	})
}

func User(c *context.Context) {
	c.JSONSuccess(&UserResp{
		c.User.ID,
		c.User.Name,
		c.User.Email,
		c.User.Admin,
		c.User.Manager,
	})
}

func CaptchaNew(c *context.Context, cpt *captcha.Captcha) {
	code, err := cpt.CreateCaptcha()
	if err != nil {
		c.JSON(500, ErrorResp{err.Error()})
		return
	}

	c.JSON(200, map[string]string{"code": code})
}

func Setup(m *macaron.Macaron) {
	m.Invoke(func(config *setting.Config) {
		reqSignOut := context.Toggle(&context.ToggleOptions{
			SignOutRequired: true,
			Config:          config,
			BaseURL:         config.GetWeb().AppSubURL})
		bindIgnErr := binding.BindIgnErr
		reqSignIn := context.Toggle(&context.ToggleOptions{
			SignInRequired: true,
			Config:         config,
			BaseURL:        config.GetWeb().AppSubURL})

		v1.Schema.GetClientRoute("auth_login").ToMacaron(m, reqSignOut, bindIgnErr(SignIn{}), Login)
		v1.Schema.GetClientRoute("auth_register").ToMacaron(m, reqSignOut, bindIgnErr(SignUp{}), Register)
		v1.Schema.GetClientRoute("auth_user").ToMacaron(m, reqSignIn, User)
		v1.Schema.GetClientRoute("auth_logout").ToMacaron(m, reqSignIn, Logout)
		v1.Schema.GetClientRoute("captcha_new").ToMacaron(m, CaptchaNew)
		v1.Schema.GetClientRoute("captcha_image").ToMacaron(m, captcha.Captchaer(captcha.Options{
			URLPrefix: config.GetWeb().BuildURI("/api/v1/client/captcha/image/"),
		}))
	})
}

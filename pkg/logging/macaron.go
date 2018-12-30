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

package logging

import (
//  context "github.com/MottainaiCI/mottainai-server/pkg/context"
  macaron "gopkg.in/macaron.v1"
  logrus "github.com/sirupsen/logrus"


	"net/http"
	"reflect"
	"time"
)
var LogTimeFormat = "2006-01-02 15:04:05"

// LoggerInvoker is an inject.FastInvoker wrapper of func(ctx *Context, log *log.Logger).
type LoggerInvoker func(ctx *macaron.Context, log *Logger)

func (invoke LoggerInvoker) Invoke(params []interface{}) ([]reflect.Value, error) {
	invoke(params[0].(*macaron.Context), params[1].(*Logger))
	return nil, nil
}

// Logger returns a middleware handler that logs the request as it goes in and the response as it goes out.
func MacaronLogger() macaron.Handler {
	return func(ctx *macaron.Context, log *Logger) {
		start := time.Now()

    log.WithFields(logrus.Fields{
      "component": "web",
      "method": ctx.Req.Method,
      "uri": ctx.Req.RequestURI,
      "from": ctx.RemoteAddr(),
      "start": start.Format(LogTimeFormat),
    }).Info("Started")

		rw := ctx.Resp.(macaron.ResponseWriter)
		ctx.Next()

    log.WithFields(logrus.Fields{
      "component": "web",
      "method": ctx.Req.Method,
      "uri": ctx.Req.RequestURI,
      "from": ctx.RemoteAddr(),
      "finish": time.Now().Format(LogTimeFormat),
      "elapsed" : time.Since(start),
      "code":rw.Status(),
      "status": http.StatusText(rw.Status()),
    }).Info("Completed")

	}
}

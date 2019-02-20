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
	"os"

	"log"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	logrus "github.com/sirupsen/logrus"
)

type Logger struct{ *logrus.Logger }

func New() *Logger { return &Logger{Logger: logrus.New()} }

func (l *Logger) SetupWithConfig(defaultLogger bool, config *setting.Config) {
	l.Setup(defaultLogger)
	if len(config.GetGeneral().LogLevel) > 0 {
		switch config.GetGeneral().LogLevel {
		case "info":
			l.SetLevel(logrus.InfoLevel)
		case "warn":
			l.SetLevel(logrus.WarnLevel)
		case "debug":
			l.SetLevel(logrus.DebugLevel)
		case "error":
			l.SetLevel(logrus.ErrorLevel)
		default:
			l.SetLevel(logrus.DebugLevel)
		}
	}
}

func (l *Logger) Setup(defaultLogger bool) {
	l.Out = os.Stdout

	// Redirect other singletons from standard log and other logrus
	// to this one.
	if defaultLogger {
		log.SetOutput(l.Writer())
		logrus.SetOutput(l.Writer())
	}
	//TODO:   log.SetLevel(log.WarnLevel)
	// TODO: Format
}

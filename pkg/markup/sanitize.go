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

package markup

import (
	"regexp"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

// Sanitizer is a protection wrapper of *bluemonday.Policy which does not allow
// any modification to the underlying policies once it's been created.
type Sanitizer struct {
	policy *bluemonday.Policy
	init   sync.Once
}

var sanitizer = &Sanitizer{
	policy: bluemonday.UGCPolicy(),
}

// NewSanitizer initializes sanitizer with allowed attributes based on settings.
// Multiple calls to this function will only create one instance of Sanitizer during
// entire application lifecycle.
func NewSanitizer() {
	sanitizer.init.Do(func() {
		// We only want to allow HighlightJS specific classes for code blocks
		sanitizer.policy.AllowAttrs("class").Matching(regexp.MustCompile(`^language-\w+$`)).OnElements("code")

		// Checkboxes
		sanitizer.policy.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
		sanitizer.policy.AllowAttrs("checked", "disabled").OnElements("input")

		// Custom URL-Schemes
		//sanitizer.policy.AllowURLSchemes(setting.Configuration.Markdown.CustomURLSchemes...)
	})
}

// Sanitize takes a string that contains a HTML fragment or document and applies policy whitelist.
func Sanitize(s string) string {
	return sanitizer.policy.Sanitize(s)
}

// SanitizeBytes takes a []byte slice that contains a HTML fragment or document and applies policy whitelist.
func SanitizeBytes(b []byte) []byte {
	return sanitizer.policy.SanitizeBytes(b)
}

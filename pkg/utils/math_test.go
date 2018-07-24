/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>

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
	"testing"
)

func TestLogisticMap(t *testing.T) {
	// Check yourself! http://www.wolframalpha.com/widgets/view.jsp?id=c731077c04035ac9e92a3706288db18f
	if LogisticMap(3.9, 0.2) != 0.6240000000000001 {
		t.Error("Logmap fails :(")
	}
	if LogisticMapSteps(4, 3.9, 0.2) != 0.8239731430433209 {
		t.Error("Unexpected 4 step result")
	}
	if FeatureScaling(20, 40, 1, 4) != 2.4615384615384617 {
		t.Error("Feature Scaling error ")
	}
}

/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
Copyright (C) 2021  Adib Saad <adib.saad@gmail.com>
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
package dbcommon

import (
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

type TaskResult struct {
	Total int               `json:"total"`
	Tasks []agenttasks.Task `json:"tasks"`
}

type TaskFilter struct {
	PageIndex int
	PageSize  int
	Sort      string
	SortOrder string
}

func CreateTaskFilter(pageIdx int, pageSize int, sort string, sortOrder string) TaskFilter {
	f := TaskFilter{
		PageIndex: pageIdx,
		PageSize:  pageSize,
		Sort:      sort,
		SortOrder: sortOrder,
	}

	if f.PageIndex < 0 {
		f.PageIndex = 0
	}

	if f.PageSize <= 0 || f.PageSize >= 50 {
		f.PageSize = 10
	}

	if _, ok := TaskSortOptions[f.Sort]; !ok {
		f.Sort = "ID"
	}

	if f.Sort == "ID" {
		f.Sort = "_key"
	}

	if _, ok := SortOrderOptions[f.SortOrder]; !ok {
		f.SortOrder = "DESC"
	}

	return f
}

var TaskSortOptions = map[string]bool{
	"ID":         true,
	"name":       true,
	"image":      true,
	"status":     true,
	"start_time": true,
}

var SortOrderOptions = map[string]bool{
	"ASC":  true,
	"DESC": true,
}

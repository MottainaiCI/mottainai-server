/*

Copyright (C) 2019  Ettore Di Giacinto <mudler@gentoo.org>
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

package v1

import (
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
)

var Schema schema.RouteGenerator = &schema.APIRouteGenerator{
	Setting: map[string]schema.Route{
		"create":   &schema.APIRoute{Path: "/api/settings", Type: "post"},
		"remove":   &schema.APIRoute{Path: "/api/settings/remove/:key", Type: "get"},
		"show_all": &schema.APIRoute{Path: "/api/settings", Type: "get"},
		"update":   &schema.APIRoute{Path: "/api/settings/update", Type: "post"},
	},
	Stats: map[string]schema.Route{
		"info": &schema.APIRoute{Path: "/api/stats", Type: "get"},
	},
	Storage: map[string]schema.Route{
		"show_all":       &schema.APIRoute{Path: "/api/storage/list", Type: "get"},
		"show_artefacts": &schema.APIRoute{Path: "/api/storage/:id/list", Type: "get"},
		"create":         &schema.APIRoute{Path: "/api/storage/:name/create", Type: "get"},
		"delete":         &schema.APIRoute{Path: "/api/storage/:id/delete", Type: "get"},
		"remove_path":    &schema.APIRoute{Path: "/api/storage/:id/remove/:path", Type: "get"},
		"show":           &schema.APIRoute{Path: "/api/storage/:id/show", Type: "get"},

		"upload": &schema.APIRoute{Path: "/api/storage/upload", Type: "post"},
	},
	Token: map[string]schema.Route{
		"show":   &schema.APIRoute{Path: "/api/token", Type: "get"},
		"create": &schema.APIRoute{Path: "/api/token/create", Type: "get"},
		"delete": &schema.APIRoute{Path: "/api/token/delete/:id", Type: "get"},
	},
	User: map[string]schema.Route{
		"show_all":      &schema.APIRoute{Path: "/api/user/list", Type: "get"},
		"show":          &schema.APIRoute{Path: "/api/user/show/:id", Type: "get"},
		"set_admin":     &schema.APIRoute{Path: "/api/user/set/admin/:id", Type: "get"},
		"unset_admin":   &schema.APIRoute{Path: "/api/user/unset/admin/:id", Type: "get"},
		"set_manager":   &schema.APIRoute{Path: "/api/user/set/manager/:id", Type: "get"},
		"unset_manager": &schema.APIRoute{Path: "/api/user/unset/manager/:id", Type: "get"},
		"delete":        &schema.APIRoute{Path: "/api/user/delete/:id", Type: "get"},

		"create": &schema.APIRoute{Path: "/api/user/create", Type: "post"},
		"edit":   &schema.APIRoute{Path: "/api/user/edit/:id", Type: "post"},
	},
	Namespace: map[string]schema.Route{
		"show_all":       &schema.APIRoute{Path: "/api/namespace/list", Type: "get"},
		"show_artefacts": &schema.APIRoute{Path: "/api/namespace/:name/list", Type: "get"},
		"create":         &schema.APIRoute{Path: "/api/namespace/:name/create", Type: "get"},
		"delete":         &schema.APIRoute{Path: "/api/namespace/:name/delete", Type: "get"},
		"tag":            &schema.APIRoute{Path: "/api/namespace/:name/tag/:taskid", Type: "get"},
		"append":         &schema.APIRoute{Path: "/api/namespace/:name/append/:taskid", Type: "get"},
		"clone":          &schema.APIRoute{Path: "/api/namespace/:name/clone/:from", Type: "get"},

		"remove": &schema.APIRoute{Path: "/api/namespace/remove", Type: "post"},
		"upload": &schema.APIRoute{Path: "/api/namespace/upload", Type: "post"},
	},
	WebHook: map[string]schema.Route{
		"show_all":        &schema.APIRoute{Path: "/api/webhook", Type: "get"},
		"create":          &schema.APIRoute{Path: "/api/webhook/create/:type", Type: "get"},
		"show":            &schema.APIRoute{Path: "/api/webhook/show/:id", Type: "get"},
		"delete":          &schema.APIRoute{Path: "/api/webhook/delete/:id", Type: "get"},
		"update_task":     &schema.APIRoute{Path: "/api/webhook/update/task/:id", Type: "post"},
		"update_pipeline": &schema.APIRoute{Path: "/api/webhook/update/pipeline/:id", Type: "post"},
		"delete_task":     &schema.APIRoute{Path: "/api/webhook/delete/task/:id", Type: "post"},
		"delete_pipeline": &schema.APIRoute{Path: "/api/webhook/delete/pipeline/:id", Type: "post"},
		"set_field":       &schema.APIRoute{Path: "/api/webhook/set", Type: "post"},
	},
	Secret: map[string]schema.Route{
		"show_all":     &schema.APIRoute{Path: "/api/secret", Type: "get"},
		"create":       &schema.APIRoute{Path: "/api/secret/create/:name", Type: "get"},
		"show":         &schema.APIRoute{Path: "/api/secret/show/:id", Type: "get"},
		"show_by_name": &schema.APIRoute{Path: "/api/secret/search/name/:name", Type: "get"},
		"delete":       &schema.APIRoute{Path: "/api/secret/delete/:id", Type: "get"},
		"set_field":    &schema.APIRoute{Path: "/api/secret/set", Type: "post"},
	},
	Node: map[string]schema.Route{
		"show_all":   &schema.APIRoute{Path: "/api/nodes", Type: "get"},
		"create":     &schema.APIRoute{Path: "/api/nodes/add", Type: "get"},
		"show":       &schema.APIRoute{Path: "/api/nodes/show/:id", Type: "get"},
		"show_tasks": &schema.APIRoute{Path: "/api/nodes/tasks/:key", Type: "get"},
		"delete":     &schema.APIRoute{Path: "/api/nodes/delete/:id", Type: "get"},

		"register": &schema.APIRoute{Path: "/api/nodes/register", Type: "post"},
	},
	Task: map[string]schema.Route{
		"show_all": &schema.APIRoute{Path: "/api/tasks", Type: "get"},
		"create":   &schema.APIRoute{Path: "/api/tasks", Type: "post"},
		"start":    &schema.APIRoute{Path: "/api/tasks/start/:id", Type: "get"},
		"clone":    &schema.APIRoute{Path: "/api/tasks/clone/:id", Type: "get"},
		"status":   &schema.APIRoute{Path: "/api/tasks/status/:status", Type: "get"},

		"stop":   &schema.APIRoute{Path: "/api/tasks/stop/:id", Type: "get"},
		"delete": &schema.APIRoute{Path: "/api/tasks/delete/:id", Type: "get"},

		"update":       &schema.APIRoute{Path: "/api/tasks/update", Type: "get"},
		"update_field": &schema.APIRoute{Path: "/api/tasks/updatefield", Type: "get"},
		"update_node":  &schema.APIRoute{Path: "/api/tasks/update/node", Type: "get"},

		"append": &schema.APIRoute{Path: "/api/tasks/append", Type: "post"},

		"as_json":       &schema.APIRoute{Path: "/api/tasks/:id", Type: "get"},
		"as_yaml":       &schema.APIRoute{Path: "/api/tasks/:id.yaml", Type: "get"},
		"stream_output": &schema.APIRoute{Path: "/api/tasks/stream_output/:id/:pos", Type: "get"},
		"tail_output":   &schema.APIRoute{Path: "/api/tasks/tail_output/:id/:pos", Type: "get"},

		"artefact_list":     &schema.APIRoute{Path: "/api/tasks/:id/artefacts", Type: "get"},
		"all_artefact_list": &schema.APIRoute{Path: "/api/artefacts", Type: "get"},

		"create_plan": &schema.APIRoute{Path: "/api/tasks/plan", Type: "post"},
		"plan_list":   &schema.APIRoute{Path: "/api/tasks/planned", Type: "get"},
		"plan_delete": &schema.APIRoute{Path: "/api/tasks/plan/delete/:id", Type: "get"},
		"plan_show":   &schema.APIRoute{Path: "/api/tasks/plan/:id", Type: "get"},

		// FIXME: Move task_log away from here
		"task_log": &schema.APIRoute{Path: "/artefact/:id/build_:id.log", Type: "get"},

		"create_pipeline":  &schema.APIRoute{Path: "/api/tasks/pipeline", Type: "post"},
		"pipeline_list":    &schema.APIRoute{Path: "/api/tasks/pipelines", Type: "get"},
		"pipeline_delete":  &schema.APIRoute{Path: "/api/tasks/pipelines/delete/:id", Type: "get"},
		"pipeline_show":    &schema.APIRoute{Path: "/api/tasks/pipeline/:id", Type: "get"},
		"pipeline_as_yaml": &schema.APIRoute{Path: "/api/tasks/pipeline/:id.yaml", Type: "get"},
		"artefact_upload":  &schema.APIRoute{Path: "/api/tasks/artefact/upload", Type: "post"},
	},
	Client: map[string]schema.Route{
		// auth
		"auth_register": &schema.APIRoute{Path: "/api/v1/client/auth/register", Type: "post"},
		"auth_login":    &schema.APIRoute{Path: "/api/v1/client/auth/login", Type: "post"},
		"auth_logout":   &schema.APIRoute{Path: "/api/v1/client/auth/logout", Type: "post"},
		"auth_user":     &schema.APIRoute{Path: "/api/v1/client/auth/user", Type: "get"},
		"captcha_new":   &schema.APIRoute{Path: "/api/v1/client/captcha/new", Type: "get"},
		"captcha_image": &schema.APIRoute{Path: "/api/v1/client/captcha/image/:id", Type: "get"},

		// dashboard
		"dashboard_stats": &schema.APIRoute{Path: "/api/v1/client/dashboard/stats", Type: "get"},
	},
}

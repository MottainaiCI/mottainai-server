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
		"create": &schema.APIRoute{
			Path:        "/api/settings",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"remove":   &schema.APIRoute{Path: "/api/settings/remove/:key", Type: "get"},
		"show_all": &schema.APIRoute{Path: "/api/settings", Type: "get"},
		"update": &schema.APIRoute{
			Path:        "/api/settings/update",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
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

		"upload": &schema.APIRoute{
			Path:        "/api/storage/upload",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
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

		"create": &schema.APIRoute{
			Path:        "/api/user/create",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"edit": &schema.APIRoute{
			Path:        "/api/user/edit/:id",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
	},
	Namespace: map[string]schema.Route{
		"show_all":       &schema.APIRoute{Path: "/api/namespace/list", Type: "get"},
		"show_artefacts": &schema.APIRoute{Path: "/api/namespace/:name/list", Type: "get"},
		"create":         &schema.APIRoute{Path: "/api/namespace/:name/create", Type: "get"},
		"delete":         &schema.APIRoute{Path: "/api/namespace/:name/delete", Type: "get"},
		"tag":            &schema.APIRoute{Path: "/api/namespace/:name/tag/:taskid", Type: "get"},
		"append":         &schema.APIRoute{Path: "/api/namespace/:name/append/:taskid", Type: "get"},
		"clone":          &schema.APIRoute{Path: "/api/namespace/:name/clone/:from", Type: "get"},

		"remove": &schema.APIRoute{
			Path:        "/api/namespace/remove",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"upload": &schema.APIRoute{
			Path:        "/api/namespace/upload",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
	},
	WebHook: map[string]schema.Route{
		"show_all": &schema.APIRoute{Path: "/api/webhook", Type: "get"},
		"create":   &schema.APIRoute{Path: "/api/webhook/create/:type", Type: "get"},
		"show":     &schema.APIRoute{Path: "/api/webhook/show/:id", Type: "get"},
		"delete":   &schema.APIRoute{Path: "/api/webhook/delete/:id", Type: "get"},
		"update_task": &schema.APIRoute{
			Path:        "/api/webhook/update/task/:id",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"update_pipeline": &schema.APIRoute{
			Path:        "/api/webhook/update/pipeline/:id",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"delete_task": &schema.APIRoute{
			Path:        "/api/webhook/delete/task/:id",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"delete_pipeline": &schema.APIRoute{
			Path:        "/api/webhook/delete/pipeline/:id",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"set_field": &schema.APIRoute{
			Path:        "/api/webhook/set",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
	},
	Secret: map[string]schema.Route{
		"show_all":     &schema.APIRoute{Path: "/api/secret", Type: "get"},
		"create":       &schema.APIRoute{Path: "/api/secret/create/:name", Type: "get"},
		"show":         &schema.APIRoute{Path: "/api/secret/show/:id", Type: "get"},
		"show_by_name": &schema.APIRoute{Path: "/api/secret/search/name/:name", Type: "get"},
		"delete":       &schema.APIRoute{Path: "/api/secret/delete/:id", Type: "get"},
		"set_field": &schema.APIRoute{
			Path:        "/api/secret/set",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
	},
	Node: map[string]schema.Route{
		"show_all":   &schema.APIRoute{Path: "/api/nodes", Type: "get"},
		"create":     &schema.APIRoute{Path: "/api/nodes/add", Type: "get"},
		"show":       &schema.APIRoute{Path: "/api/nodes/show/:id", Type: "get"},
		"show_tasks": &schema.APIRoute{Path: "/api/nodes/tasks/:key", Type: "get"},
		"delete":     &schema.APIRoute{Path: "/api/nodes/delete/:id", Type: "get"},

		"register": &schema.APIRoute{
			Path:        "/api/nodes/register",
			Type:        "post",
			ContentType: schema.ContentTypeJson,
		},
	},
	NodeQueue: map[string]schema.Route{
		"show_all": &schema.APIRoute{Path: "/api/nodequeues", Type: "get"},
		"create": &schema.APIRoute{
			Path:        "/api/nodequeues/add",
			Type:        "post",
			ContentType: schema.ContentTypeJson,
		},
		"add_task": &schema.APIRoute{
			Path:        "/api/nodequeues/addtask/:queue/:tid",
			Type:        "put",
			ContentType: schema.ContentTypeJson,
		},
		"del_task": &schema.APIRoute{
			Path:        "/api/nodequeues/deltask/:queue/:tid",
			Type:        "delete",
			ContentType: schema.ContentTypeJson,
		},
		"delete": &schema.APIRoute{
			Path:        "/api/nodequeues/delete",
			Type:        "delete",
			ContentType: schema.ContentTypeJson,
		},
		"show": &schema.APIRoute{
			Path:        "/api/nodequeues/show/:id",
			Type:        "get",
			ContentType: schema.ContentTypeJson,
		},
		"show_byagent": &schema.APIRoute{
			Path:        "/api/nodequeues/shownode/:nodeid",
			Type:        "get",
			ContentType: schema.ContentTypeJson,
		},
	},
	Queue: map[string]schema.Route{
		"show_all": &schema.APIRoute{Path: "/api/queues", Type: "get"},
		"create": &schema.APIRoute{
			Path:        "/api/queues/add/:name",
			Type:        "post",
			ContentType: schema.ContentTypeJson,
		},
		"delete": &schema.APIRoute{
			Path: "/api/queues/:qid/delete",
			Type: "delete",
		},
		"show": &schema.APIRoute{
			Path:        "/api/queues/:qid/show",
			Type:        "get",
			ContentType: schema.ContentTypeJson,
		},
		"add_task_in_progress": &schema.APIRoute{
			Path: "/api/queues/:qid/task-in-progress/:tid",
			Type: "post",
		},
		"del_task_in_progress": &schema.APIRoute{
			Path: "/api/queues/:qid/task-in-progress/:tid",
			Type: "delete",
		},
		"add_task": &schema.APIRoute{
			Path: "/api/queues/:qid/task/:tid",
			Type: "post",
		},
		"del_task": &schema.APIRoute{
			Path: "/api/queues/:qid/task/:tid",
			Type: "delete",
		},
		"get_qid": &schema.APIRoute{
			Path:        "/api/queues/:name",
			Type:        "get",
			ContentType: schema.ContentTypeJson,
		},
		"add_pipeline_in_progress": &schema.APIRoute{
			Path: "/api/queues/:qid/pipeline-in-progress/:pid",
			Type: "post",
		},
		"del_pipeline_in_progress": &schema.APIRoute{
			Path: "/api/queues/:qid/pipeline-in-progress/:pid",
			Type: "delete",
		},
		"add_pipeline": &schema.APIRoute{
			Path: "/api/queues/:qid/pipeline/:pid",
			Type: "post",
		},
		"del_pipeline": &schema.APIRoute{
			Path: "/api/queues/:qid/pipeline/:pid",
			Type: "delete",
		},
		"reset": &schema.APIRoute{
			Path: "/api/queues/:qid/reset",
			Type: "post",
		},
	},
	Task: map[string]schema.Route{
		"show_all":          &schema.APIRoute{Path: "/api/tasks", Type: "get"},
		"show_all_filtered": &schema.APIRoute{Path: "/api/tasks_filtered", Type: "get"},
		"create": &schema.APIRoute{
			Path:        "/api/tasks",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"start":  &schema.APIRoute{Path: "/api/tasks/start/:id", Type: "get"},
		"clone":  &schema.APIRoute{Path: "/api/tasks/clone/:id", Type: "get"},
		"status": &schema.APIRoute{Path: "/api/tasks/status/:status", Type: "get"},

		"stop":   &schema.APIRoute{Path: "/api/tasks/stop/:id", Type: "get"},
		"delete": &schema.APIRoute{Path: "/api/tasks/delete/:id", Type: "get"},

		"update":       &schema.APIRoute{Path: "/api/tasks/update", Type: "get"},
		"update_field": &schema.APIRoute{Path: "/api/tasks/updatefield", Type: "get"},
		"update_node":  &schema.APIRoute{Path: "/api/tasks/update/node", Type: "get"},

		"append": &schema.APIRoute{
			Path:        "/api/tasks/append",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},

		"as_json":       &schema.APIRoute{Path: "/api/tasks/:id", Type: "get"},
		"as_yaml":       &schema.APIRoute{Path: "/api/tasks/:id.yaml", Type: "get"},
		"stream_output": &schema.APIRoute{Path: "/api/tasks/stream_output/:id/:pos", Type: "get"},
		"tail_output":   &schema.APIRoute{Path: "/api/tasks/tail_output/:id/:pos", Type: "get"},

		"artefact_list":     &schema.APIRoute{Path: "/api/tasks/:id/artefacts", Type: "get"},
		"all_artefact_list": &schema.APIRoute{Path: "/api/artefacts", Type: "get"},

		"create_plan": &schema.APIRoute{
			Path:        "/api/tasks/plan",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"plan_list":   &schema.APIRoute{Path: "/api/tasks/planned", Type: "get"},
		"plan_delete": &schema.APIRoute{Path: "/api/tasks/plan/delete/:id", Type: "get"},
		"plan_show":   &schema.APIRoute{Path: "/api/tasks/plan/:id", Type: "get"},

		// FIXME: Move task_log away from here
		"task_log": &schema.APIRoute{Path: "/artefact/:id/build_:id.log", Type: "get"},

		"create_pipeline": &schema.APIRoute{
			Path:        "/api/tasks/pipeline",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"pipeline_list":    &schema.APIRoute{Path: "/api/tasks/pipelines", Type: "get"},
		"pipeline_delete":  &schema.APIRoute{Path: "/api/tasks/pipelines/delete/:id", Type: "get"},
		"pipeline_show":    &schema.APIRoute{Path: "/api/tasks/pipeline/:id", Type: "get"},
		"pipeline_as_yaml": &schema.APIRoute{Path: "/api/tasks/pipeline/:id.yaml", Type: "get"},
		"pipeline_completed": &schema.APIRoute{
			Path: "/api/tasks/pipeline/:id/completed",
			Type: "post",
		},

		"artefact_upload": &schema.APIRoute{
			Path:        "/api/tasks/artefact/upload",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
	},
	Callbacks: map[string]schema.Route{
		"cb_int_gh": &schema.APIRoute{Path: "/callbacks/integrations/github", Type: "get"},
	},
	Client: map[string]schema.Route{
		// auth
		"auth_register": &schema.APIRoute{
			Path:        "/api/v1/client/auth/register",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"auth_login": &schema.APIRoute{
			Path:        "/api/v1/client/auth/login",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"auth_logout": &schema.APIRoute{
			Path:        "/api/v1/client/auth/logout",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"auth_int_github": &schema.APIRoute{
			Path:        "/api/v1/client/auth/int/github",
			Type:        "get",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"auth_int_github_callback": &schema.APIRoute{
			Path:        "/api/v1/client/auth/int/github_callback",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"auth_int_github_logout": &schema.APIRoute{
			Path:        "/api/v1/client/auth/int/github_logout",
			Type:        "post",
			ContentType: schema.ContentTypeFormUrlEncoded,
		},
		"auth_user":     &schema.APIRoute{Path: "/api/v1/client/auth/user", Type: "get"},
		"captcha_new":   &schema.APIRoute{Path: "/api/v1/client/captcha/new", Type: "get"},
		"captcha_image": &schema.APIRoute{Path: "/api/v1/client/captcha/image/:id", Type: "get"},

		// dashboard
		"dashboard_stats": &schema.APIRoute{Path: "/api/v1/client/dashboard/stats", Type: "get"},

		// users
		"users_show_all": &schema.APIRoute{Path: "/api/v1/client/users/list", Type: "get"},
		"users_show":     &schema.APIRoute{Path: "/api/v1/client/users/show/:id", Type: "get"},
		"users_create":   &schema.APIRoute{Path: "/api/v1/client/users/create", Type: "post"},
		"users_delete":   &schema.APIRoute{Path: "/api/v1/client/users/delete/:id", Type: "post"},
		"users_edit":     &schema.APIRoute{Path: "/api/v1/client/users/edit/:id", Type: "post"},
	},
}

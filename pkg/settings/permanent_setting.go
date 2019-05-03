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

package setting

const SYSTEM_SIGNUP_ENABLED = "system.signup"
const SYSTEM_WEBHOOK_ENABLED = "system.webhook"
const SYSTEM_THIRDPARTY_INTEGRATION_ENABLED = "system.thirdparty_integration"
const SYSTEM_WEBHOOK_PR_ENABLED = "system.webhook.pull_request"
const SYSTEM_WEBHOOK_INTERNAL_ONLY = "system.webhook.internal_only"
const SYSTEM_WEBHOOK_DEFAULT_QUEUE = "system.webhook.default_queue"

const SYSTEM_PROTECT_NAMESPACE_OVERWRITE = "system.namespace.protect_overwrite"
const SYSTEM_PROTECT_NAMESPACE_PARALLEL_APPEND = "system.namespace.protect_overwrite.parallel_append"

// This option could be used when user/password verification is
// done from external compoenent or reverse proxy and permit
// to validate only user and ignore password
const SYSTEM_SIGNIN_ONLY_USERVALIDATION = "system.signin_useronly_validation"

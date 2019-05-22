#!/bin/bash
#

test_simple_task() {

cat <<EOF > /tmp/task_01.yaml
# Task name
name: "My task"

# Image used by the task player
image: "sabayon/base-amd64"

# Script to be executed
script:
  - echo test > artefacts/hello

# Task type
type: docker
tag_namespace: "test1"
EOF


  result="$(mottainai-wrapper task create --yaml /tmp/task_01.yaml)"
  assertContains 'Task has been created' "${result}" "has been created"

  mottainai-wrapper task create --yaml /tmp/task_01.yaml --monitor
  assertEquals 'Finishes successfully' $? 0

  mottainai-wrapper namespace download test1 /tmp
  assertContains 'Artifacts has been downloaded correctly' "$(cat /tmp/hello)" "test"

  agent_log="$(docker logs mott-test)"
  assertContains 'Task has been executed' "${agent_log}" "Context RootTaskDir"
}



# Load and run shUnit2.
[ -n "${ZSH_VERSION:-}" ] && SHUNIT_PARENT=$0
. ../shunit2

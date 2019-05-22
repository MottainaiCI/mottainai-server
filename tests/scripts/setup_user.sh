#!/bin/bash

set -xe
# test:test credentials
curl -X POST --data-binary @- --dump - http://localhost:8529/_db/mottainai/_api/document/Users <<'EOF'
{
  "_key": "348",
  "email": "test@test.com",
  "identities": null,
  "is_admin": "yes",
  "is_manager": "",
  "name": "test",
  "password": "$s2$16384$8$1$xz49ynQhM2JlCMDTZJd4gBca$AUvkpqbteYuTIhxes27ynssG1beBib2ujjXHACqo70Q="
}
EOF

curl -X POST --data-binary @- --dump - http://localhost:8529/_db/mottainai/_api/document/Tokens <<'EOF'
{
  "key": "8d9439805bd4d32633d8fae3ed2375e0fc7f35a591ecdf2880c111e32a77361b",
  "user_id": "348"
}
EOF

curl -X POST --data-binary @- --dump - http://localhost:8529/_db/mottainai/_api/document/Nodes <<'EOF'
{
    "key": "8d9439805bd4d32633d8f",
    "owner": "348"
}

EOF


#!/bin/bash
set -ex

SCRIPT_DIR="${SCRIPT_DIR:-tests/scripts}"
WORKDIR="${WORKDIR:-$PWD}"
SUBPROJECT="${SUBPROJECT:-mottainai-server}"
export CHECKOUT_AGENT="${CHECKOUT_AGENT:-master}"


export GOPATH=/tmp/go
export PATH=$PATH:$GOPATH/bin
export VENDORDIR=github.com/MottainaiCI
mkdir -p ${GOPATH}/src/${VENDORDIR}
echo "[*] Installing deps, may take a while.."
sudo ACCEPT_LICENSE=* equo i git dev-lang/go enman docker docker-compose 2>&1 >/dev/null
sudo systemctl start docker
sudo chmod 777 /var/run/docker.sock
sudo enman add https://dispatcher.sabayon.org/sbi/namespace/devel/devel 2>&1 >/dev/null
sudo equo up 2>&1 >/dev/null
sudo equo i mottainai-cli 2>&1 >/dev/null

if [ -n "$CHECKOUT" ]; then 
    git clone https://${VENDORDIR}/${SUBPROJECT} ${GOPATH}/src/${VENDORDIR}/${SUBPROJECT}
    pushd ${GOPATH}/src/${VENDORDIR}/${SUBPROJECT} >/dev/null
        git checkout $CHECKOUT
    popd >/dev/null 
else
    mkdir -p ${GOPATH}/src/${VENDORDIR}/${SUBPROJECT} || true
    cp -rf $WORKDIR/* ${GOPATH}/src/${VENDORDIR}/${SUBPROJECT}
fi

sudo cp -rf ${GOPATH}/src/${VENDORDIR}/${SUBPROJECT}/${SCRIPT_DIR}/mcliwrap.sh /usr/bin/mottainai-wrapper
sudo chmod a+x /usr/bin/mottainai-wrapper

pushd ${GOPATH}/src/${VENDORDIR}/${SUBPROJECT} >/dev/null
    go build
    sudo cp -rf mottainai-server /usr/bin/mottainai-server
    tmpdir=`mktemp --tmpdir -d`
    cp -rf contrib/docker-compose "$tmpdir"
    cp -rf mottainai-server "$tmpdir/docker-compose/"

    pushd "$tmpdir/docker-compose" >/dev/null
        mv docker-compose.arangodb.yml docker-compose.yml
        sed -i "s|#- ./mottainai-server.yaml:/etc/mottainai/mottainai-server.yaml|- "$PWD"/mottainai-server:/usr/bin/mottainai-server|g" docker-compose.yml
        sed -i "s|# For static config:|- "$PWD":/var/lib/mottainai|g" docker-compose.yml
        docker-compose up -d
    popd >/dev/null
popd >/dev/null

git clone https://${VENDORDIR}/mottainai-agent ${GOPATH}/src/${VENDORDIR}/mottainai-agent
go get github.com/spf13/viper
go get github.com/spf13/cobra
rm -rf ${GOPATH}/src/${VENDORDIR}/mottainai-server/vendor/github.com/spf13 || true

pushd ${GOPATH}/src/${VENDORDIR}/mottainai-agent >/dev/null
    git checkout $CHECKOUT_AGENT
    rm -rf ${GOPATH}/src/${VENDORDIR}/mottainai-agent/vendor/github.com/spf13 || true
    rm -rf ${GOPATH}/src/${VENDORDIR}/mottainai-agent/vendor/github.com/MottainaiCI/mottainai-server || true
    go build
    sudo cp -rf mottainai-agent /usr/bin/mottainai-agent
    sudo chmod a+x /usr/bin/mottainai-agent
popd >/dev/null

sleep 120

# FIXME: https://github.com/docker/compose/issues/3270#issuecomment-206214034
docker exec -u 0 docker-compose_mottainai_1 /bin/bash -c 'chown -R mottainai-server:mottainai /srv/mottainai/web/'

bash -ex ${GOPATH}/src/${VENDORDIR}/${SUBPROJECT}/${SCRIPT_DIR}/setup_user.sh
IP=$(ifconfig -a | awk '/(cast)/ { print $2 }' | cut -d':' -f2 | head -1)

# Prepare one node

docker run -d -v /var/run/docker.sock:/var/run/docker.sock \
	   -v /usr/bin/mottainai-agent:/usr/bin/mottainai-agent \
	   -e MOTTAINAI_AGENT_WEB__APPLICATION_URL=http://$IP:4545 \
	   -e MOTTAINAI_AGENT_BROKER__TYPE=redis \
	   -e MOTTAINAI_AGENT_BROKER__BROKER=redis://$IP:6379/1 \
	   -e MOTTAINAI_AGENT_BROKER__RESULT_BACKEND=redis://$IP:6379/2 \
	   -e MOTTAINAI_AGENT_BROKER__DEFAULT_QUEUE=standard \
	   -e MOTTAINAI_AGENT_BROKER__EXCHANGE=jobs \
	   -e MOTTAINAI_AGENT_AGENT__AGENT_KEY=8d9439805bd4d32633d8f \
	   -e MOTTAINAI_AGENT_AGENT__API_KEY=8d9439805bd4d32633d8fae3ed2375e0fc7f35a591ecdf2880c111e32a77361b \
	   -v /srv/mottainai/build:/srv/mottainai/build --name mott-test --rm mottainaici/agent agent

sleep 60

docker logs mott-test

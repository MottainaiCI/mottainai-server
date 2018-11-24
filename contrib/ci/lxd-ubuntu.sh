#!/bin/bash
# Author: Daniele Rondina <geaaru@sabayonlinux.org>
# Description: Script to execute mottainai test and build on ubuntu container.

set -e

CONTAINER_IMAGE=${CONTAINER_IMAGE:-ubuntu:16.04}
CONTAINER_NAME=${CONTAINER_NAME:-mottainai-ubuntu-test}
# By default use ephemerl container
LXD_LAUNCH_OPTS=${LXD_LAUNCH_OPTS:--e}

# Container receive DHCP address or we need configure static address.
CONTAINER_NET_STATIC=${CONTAINER_NET_STATIC:-0}
CONTAINER_NET_ADDR=${CONTAINER_NET_ADDR:-192.168.20.10/24}
CONTAINER_NET_IFACE=${CONTAINER_NET_IFACE:-eth0}
CONTAINER_NET_GW=${CONTAINER_NET_GW:-192.168.20.100}

GOLANG_VERSION=${GOLANG_VERSION:-1.10}

# Ubuntu 14 doesn't work. I don't find repository with golang-race-detector-runtime
# apt-get install -y gcc-arm-linux-gnueabi libc6-dev-armhf-cross libc6-dev golang-${GOLANG_VERSION}-go gccgo gccgo-go gcc-multilib

# apt-get install -y gccgo-arm-linux-gnueabihf gccgo-4.7-arm-linux-gnueabi gobjc-4.8-multilib-arm-linux-gnueabihf

SCRIPT_COMMANDS=$(cat <<-END
#!/bin/bash

set -e

echo 'nameserver 1.1.1.1' > /etc/resolv.conf

sleep 1

echo "deb http://archive.ubuntu.com/ubuntu xenial-backports main universe multiverse restricted" >> /etc/apt/sources.list

apt-get update

apt-get install -y libc6-dev-armhf-armel-cross libc6-dev golang-${GOLANG_VERSION}-go
apt-get install -y libc6-dev-i386 libc6-dev-armhf-armel-cross linux-headers-generic
apt-get install -y git make
apt-get install -y golang-${GOLANG_VERSION}-race-detector-runtime
apt-get install -y gcc-arm-linux-gnueabi

cp --archive /usr/include/asm-generic /usr/include/asm

mkdir ~/go/src/github.com/MottainaiCI -p

cd ~/go/src/github.com/MottainaiCI

git clone https://github.com/MottainaiCI/mottainai-server.git

cd mottainai-server

git checkout lxd-integration

PATH=/usr/lib/go-${GOLANG_VERSION}/bin:\$PATH make deps
PATH=/usr/lib/go-${GOLANG_VERSION}/bin:\$HOME/go/bin:\$PATH make multiarch-build EXTENSIONS=lxd
PATH=/usr/lib/go-${GOLANG_VERSION}/bin:\$HOME/go/bin:\$PATH make build-test EXTENSIONS=lxd
END
)


container_exec () {
  lxc exec ${CONTAINER_NAME} $@
}

container_config_net () {
  if [ "${CONTAINER_NET_STATIC}" = 1 ] ; then
    container_exec ip a a ${CONTAINER_NET_ADDR} dev ${CONTAINER_NET_IFACE}
    container_exec ip r a default via ${CONTAINER_NET_GW}
  fi
}

main () {
  # Create container
  lxc launch ${LXD_LAUNCH_OPTS} ${CONTAINER_IMAGE} ${CONTAINER_NAME}

  container_config_net

  # Create scripts with command
  echo "${SCRIPT_COMMANDS}" > /tmp/test_mottainai.sh

  chmod a+x /tmp/test_mottainai.sh

  lxc file push /tmp/test_mottainai.sh ${CONTAINER_NAME}/tmp/

  rm /tmp/test_mottainai.sh

  container_exec /tmp/test_mottainai.sh

  # TODO: handle this as defer operation
  lxc stop ${CONTAINER_NAME}
}

main

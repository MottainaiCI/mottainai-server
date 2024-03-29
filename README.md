# Mottainai Continuous Integration Server
[![LICENSE](https://img.shields.io/badge/license-GPL%20(%3E%3D3)-blue.svg)](https://spdx.org/licenses/GPL-3.0-or-later.html) [![master](https://img.shields.io/travis/MottainaiCI/mottainai-server/master.svg?label=master)](https://travis-ci.org/MottainaiCI/mottainai-server) [![develop](https://img.shields.io/travis/MottainaiCI/mottainai-server/master.svg?label=develop)](https://travis-ci.org/MottainaiCI/mottainai-server)

Build powerful, flexible and decentralized pipelines playable locally. Manage, Publish and release your task's produced content.

Mottainai is a Task Distributed, Continous Integration and Delivery system - it allows you to build, test, deploy and manage content built from custom tasks from different nodes in a network. You can hook specific tasks to Git repositories in a CI style, or either directly execute pipelines or production tasks in safe environments.

It uses polling and queues to handle tasks and pipelines through REST API. This permits to have agents under proxy or NAT without problems.

It was developed for building and releasing packages for Linux Distributions, it is used by [Sabayon Linux](https://www.sabayon.org/) to produce and orchestrate builds of community repositories and to build LiveCDs - but it's suitable for every workflow which is artefact-oriented. 

For more information, see the [documentation](https://mottainaici.github.io/docs/), also see the [Official Gentoo project](https://wiki.gentoo.org/wiki/Project:Build_Service) page.



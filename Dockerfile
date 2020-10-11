FROM sabayon/base-amd64:latest

ENV ACCEPT_LICENSE=*

RUN equo install enman && \
    enman add https://downloads.svc.sabayon.org/devel/devel && \
    equo up && equo u && equo i mottainai-server && equo cleanup

# See: https://github.com/docker/compose/issues/3270#issuecomment-206214034
RUN mkdir -p /srv/mottainai/web/db
RUN mkdir -p /srv/mottainai/web/artefact
RUN mkdir -p /srv/mottainai/web/namespaces
RUN mkdir -p /srv/mottainai/web/storage
RUN mkdir -p /build
RUN chown -R mottainai-server:mottainai /srv/mottainai/web
RUN chown -R mottainai-server:mottainai /build


EXPOSE 9090

USER mottainai-server

VOLUME ["/etc/mottainai", "/srv/mottainai"]

ENTRYPOINT [ "/usr/bin/mottainai-server" ]

FROM quay.io/mocaccino/extra

RUN luet install -y repository/mocaccino-os-commons-stable repository/geaaru
RUN luet install -y dev-util/mottainai-server && luet cleanup

# See: https://github.com/docker/compose/issues/3270#issuecomment-206214034
RUN mkdir -p /srv/mottainai/web/db
RUN mkdir -p /srv/mottainai/web/artefact
RUN mkdir -p /srv/mottainai/web/namespaces
RUN mkdir -p /srv/mottainai/web/storage
RUN mkdir -p /build
RUN chown -R mottainai-server:mottainai /srv/mottainai/
RUN chown -R mottainai-server:mottainai /build
# Fix temporary until is fixed to upper layer
RUN chmod -R a+rwx /tmp


EXPOSE 9090

USER mottainai-server

VOLUME ["/etc/mottainai", "/srv/mottainai"]

ENTRYPOINT [ "/usr/bin/mottainai-server" ]

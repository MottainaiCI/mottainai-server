FROM sabayon/base-amd64:latest

ENV ACCEPT_LICENSE=*

RUN equo install enman && \
    enman add https://dispatcher.sabayon.org/sbi/namespace/devel/devel && \
    equo up && equo u && equo i mottainai-server && equo cleanup

# See: https://github.com/docker/compose/issues/3270#issuecomment-206214034
RUN chown -R mottainai-server:mottainai /srv/mottainai/web

EXPOSE 9090

USER mottainai-server

VOLUME ["/etc/mottainai", "/srv/mottainai"]

ENTRYPOINT [ "/usr/bin/mottainai-server" ]

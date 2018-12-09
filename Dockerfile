FROM sabayon/base-amd64:latest

ENV ACCEPT_LICENSE=*

RUN equo install enman && \
    enman add devel && \
    equo up && equo u && equo i mottainai-server

EXPOSE 9090

USER mottainai-server

VOLUME ["/etc/mottainai", "/srv/mottainai"]

ENTRYPOINT [ "/usr/bin/mottainai-server" ]

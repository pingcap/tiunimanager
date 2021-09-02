FROM pingcap/alpine-glibc
WORKDIR /usr/local/tiem
RUN mkdir bin \
    && mkdir etc \
    && mkdir logs \
    && mkdir scripts \
    && mkdir docs
COPY bin/tiupcmd ./bin
COPY bin/brcmd ./bin
COPY bin/openapi-server ./bin
COPY bin/cluster-server ./bin
COPY bin/metadb-server ./bin
COPY build_helper/docker_start_cmd.sh ./scripts
EXPOSE 4116
ENTRYPOINT ["/bin/bash", "/usr/local/tiem/scripts/docker_start_cmd.sh"]
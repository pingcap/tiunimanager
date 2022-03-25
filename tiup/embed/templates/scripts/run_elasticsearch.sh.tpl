#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}

cd "${DEPLOY_DIR}" || exit 1

LANG=en_US.UTF-8 \
exec env ES_PATH_CONF="${DEPLOY_DIR}/conf/" \
    ES_TMPDIR="{{.DataDir}}" \
    ES_JAVA_OPTS="-Xms{{.HeapSize}} -Xmx{{.HeapSize}}" \
    bin/bin/elasticsearch \
    -d

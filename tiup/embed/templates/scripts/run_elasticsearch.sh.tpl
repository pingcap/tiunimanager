#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}

cd "${DEPLOY_DIR}" || exit 1

LANG=en_US.UTF-8 \
exec env ES_PATH_CONF="${DEPLOY_DIR}/conf/" bin/bin/elasticsearch \
    -d

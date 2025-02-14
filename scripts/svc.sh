#!/bin/bash

SVC_NAME="valkyrie.service"
SVC_NAME=${SVC_NAME// /_}
SVC_DESCRIPTION="Valkyrie Aggregator Service"

SVC_CMD=$1
arg_2=${2}

RUNNER_ROOT=$(pwd)

UNIT_PATH=/etc/systemd/system/${SVC_NAME}
TEMPLATE_PATH=./valkyrie.service.template
TEMP_PATH=./valkyrie.service.temp
CONFIG_PATH=.service

user_id=$(id -u)

# systemctl must run as sudo
# this script is a convenience wrapper around systemctl
if [ "${user_id}" -ne 0 ]; then
    echo "Must run as sudo"
    exit 1
fi

function failed()
{
   local error=${1:-Undefined error}
   echo "Failed: $error" >&2
   exit 1
}

if [ ! -f "${TEMPLATE_PATH}" ]; then
    failed "Must run from package root or install is corrupt"
fi

#check if we run as root
if [[ $(id -u) != "0" ]]; then
    echo "Failed: This script requires to run with sudo." >&2
    exit 1
fi

function install()
{
    echo "Creating launch valkyrie in ${UNIT_PATH}"
    if [ -f "${UNIT_PATH}" ]; then
        failed "error: exists ${UNIT_PATH}"
    fi

    if [ -f "${TEMP_PATH}" ]; then
      rm "${TEMP_PATH}" || failed "failed to delete ${TEMP_PATH}"
    fi

    # can optionally use username supplied
    run_as_user=${arg_2:-$SUDO_USER}
    echo "Run as user: ${run_as_user}"

    run_as_uid=$(id -u "${run_as_user}") || failed "User does not exist"
    echo "Run as uid: ${run_as_uid}"

    run_as_gid=$(id -g "${run_as_user}") || failed "Group not available"
    echo "gid: ${run_as_gid}"

    sed "s/{{User}}/${run_as_user}/g; s/{{Description}}/$(echo "${SVC_DESCRIPTION}" | sed -e 's/[\/&]/\\&/g')/g; s/{{RunnerRoot}}/$(echo "${RUNNER_ROOT}" | sed -e 's/[\/&]/\\&/g')/g;" "${TEMPLATE_PATH}" > "${TEMP_PATH}" || failed "failed to create replacement temp file"
    mv "${TEMP_PATH}" "${UNIT_PATH}" || failed "failed to copy unit file"

    # unit file should not be executable and world writable
    chmod 664 "${UNIT_PATH}" || failed "failed to set permissions on ${UNIT_PATH}"
    systemctl daemon-reload || failed "failed to reload daemons"

    # Since we started with sudo, runsvc.sh will be owned by root. Change this to current login user.
    chown "${run_as_uid}:${run_as_gid}" ./valkyrie || failed "failed to set owner for valkyrie"
    chmod 755 ./valkyrie || failed "failed to set permission for valkyrie"

    systemctl enable "${SVC_NAME}" || failed "failed to enable ${SVC_NAME}"

    echo "${SVC_NAME}" > ${CONFIG_PATH} || failed "failed to create .service file"
    chown "${run_as_uid}:${run_as_gid}" "${CONFIG_PATH}" || failed "failed to set permission for ${CONFIG_PATH}"
}

function start()
{
    systemctl start "${SVC_NAME}" || failed "failed to start ${SVC_NAME}"
    status
}

function stop()
{
    systemctl stop "${SVC_NAME}" || failed "failed to stop ${SVC_NAME}"
    status
}

function uninstall()
{
    if service_exists; then
        stop
        systemctl disable "${SVC_NAME}" || failed "failed to disable ${SVC_NAME}"
        rm "${UNIT_PATH}" || failed "failed to delete ${UNIT_PATH}"
    else
        echo "Service ${SVC_NAME} is not installed"
    fi
    if [ -f "${CONFIG_PATH}" ]; then
      rm "${CONFIG_PATH}" || failed "failed to delete ${CONFIG_PATH}"
    fi
    systemctl daemon-reload || failed "failed to reload daemons"
}

function service_exists() {
    if [ -f "${UNIT_PATH}" ]; then
        return 0
    else
        return 1
    fi
}

function status()
{
    if service_exists; then
        echo
        echo "${UNIT_PATH}"
    else
        echo
        echo "not installed"
        echo
        exit 1
    fi

    systemctl --no-pager status "${SVC_NAME}"
}

function usage()
{
    echo
    echo Usage:
    echo "./svc.sh [install, start, stop, status, uninstall]"
    echo "Commands:"
    echo "   install [user]: Install valkyrie service as Root or specified user."
    echo "   start: Manually start the valkyrie service."
    echo "   stop: Manually stop the valkyrie service."
    echo "   status: Display status of valkyrie service."
    echo "   uninstall: Uninstall valkyrie service."
    echo
}

case $SVC_CMD in
   "install") install;;
   "status") status;;
   "uninstall") uninstall;;
   "start") start;;
   "stop") stop;;
   *) usage;;
esac

exit 0

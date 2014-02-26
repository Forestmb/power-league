#!/bin/bash
set -e

cd "$(dirname "$0")"

dir="."
log_dir="log"
stdout_file="server.stdout.log"
stderr_file="server.stderr.log"
binary="power-league"
pid="server.pid"
conf="server.conf"

prog="$(basename "$0")"
function usage()
{
    cat 1>&2 << EOF
Usage: ${prog} start|restart [ arg ] ...
       ${prog} stop
       ${prog} status

Options:
    [ arg ]
        If the command is 'start' or 'restart', any additional arguments
        will be passed to the server.
EOF
    exit 1
}

function error()
{
    echo "$@" 1>&2
}

function start()
{
    if [ -f "${pid}" ]
    then
        error "PID file '${pid}' already exists. Server may already be running"
        exit 1
    fi
    mkdir -p "${log_dir}"
    if [ -f "${log_dir}/${stdout_file}" ]
    then
        mv -f "${log_dir}/${stdout_file}" "${log_dir}/${stdout_file}.previous"
    fi
    if [ -f "${log_dir}/${stderr_file}" ]
    then
        mv -f "${log_dir}/${stderr_file}" "${log_dir}/${stderr_file}.previous"
    fi

    # Configure this instance of the application
    . "${conf}"
    args="${args} ${server_args}"
    args="${args} --clientKey ${client_key}"
    args="${args} --clientSecret ${client_secret}"
    if [ ! -z "${cookie_auth_key}" ]
    then
        args="${args} --cookieAuthKey ${cookie_auth_key}"
    fi
    if [ ! -z "${cookie_encryption_key}" ]
    then
        args="${args} --cookieEncryptionKey ${cookie_encryption_key}"
    fi

    nohup "${dir}/${binary}" ${args} "$@" \
            > "${log_dir}/${stdout_file}" \
            2> "${log_dir}/${stderr_file}" &

    echo "$!" > "${pid}"
    echo "Server started"
}

function stop()
{
    if [ ! -f "${pid}" ]
    then
        error "PID file '${pid}' does not exist. Server will not be stopped"
        exit 1
    fi
    
    kill "$(cat "${pid}")" 
    rm "${pid}"
    echo "Server stopped"
}

function status()
{
    if [ -f "${pid}" ]
    then
        process="$(cat "${pid}")"
        if [ ! -d "/proc/${process}/" ]
        then
            error "PID file '${pid}' exists but process with id '${process}' could not be found."
            error "Check '${log_dir}/${log_file}' for more information."
            exit 1
        else
            echo "Server is running. PID=${process}"
            exit 0
        fi
    else
        echo "Server is stopped."
        exit 0
    fi
}

case "$1" in
    start)
        shift
        start "$@"
        ;;
    stop)
        stop
        ;;
    restart|reload)
        shift
        stop
        start "$@"
        ;;
    status)
        status
        ;;
    *)
        usage
        ;;
esac

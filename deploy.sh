#!/bin/bash
set -e

cd "$(dirname "$(readlink -f "$0")")"
prog="$(basename "$0")"
function usage()
{
    cat 1>&2 << EOF
Usage: ${prog} [ options ] [ application ] [ host ]

Description: Deploy a instance of the application to a remote host.

Options:
    -h
        Display this help.
    -d <deploy-dir>
        Directory on the remote host to deploy the application to. Defaults
        to ~/power-league
    -v
        Print extra debug information.

Arguments:
    [ application ]
        The application tarball to deploy
    [ host ]
        The remote host to deploy the application to.
EOF
    exit "${1:-1}"
}

deploydir="~/power-league"
verbose="false"
while getopts ":d:hv" option; do
    case "${option}" in
        d)
            deploydir="${OPTARG}"
            ;;
        h)
            usage 0
            ;;
        v)
            verbose="true"
            ;;
        \?) echo "Unknown option: -${OPTARG}" 1>&2
            usage 1
            ;;
        *)
            break
            ;;
    esac
done
shift $((OPTIND-1))

function log_verbose
{
    if [ "${verbose}" == "true" ]
    then
        echo "$@"
    fi
}

app_file="$1"
if [ -z "${app_file}" ]
then
    echo "No application specified." 1>&2
    usage 2
fi
app_name="$(basename "${app_file}" | sed 's/\.tar\.gz$//')"

host="$2"
if [ -z "${host}" ]
then
    echo "No host specified." 1>&2
    usage 3
fi

log_verbose "Deploy dir: ${deploydir}"

link="current"
prev="previous"
old="old"
server="server.sh"
dist="build/dist"

# 1. Copies package
# 2. Updates symbolic links to latest and previous versions
# 3. Stops existing server if running
# 4. Starts new server
function server_commands()
{
    cat << EOF
#!/bin/sh
set -e

cd ${deploydir}

# If this verison is running, just replace it
if [ \$(basename "\$(readlink -f "${link}")") == "${app_name}" ]
then
    "${app_name}/${server}" stop || true
    rm -rf "${app_name}"
    tar xf "${app_name}.tar.gz"
# Otherwise, rotate
else
    mkdir -p "${old}"
    if [ -h "${prev}" ]
    then
        rm -rf "${old}/\$(basename "\$(readlink -f "${prev}")")"
        mv -f "\$(readlink -f "${prev}")" "${old}"
    fi

    tar xf "${app_name}.tar.gz"

    if [ -h "${link}" ]
    then
        ln -s -f -T \`readlink -f "${link}" \` "${prev}"
    fi
    ln -s -f -T "\`pwd\`/${app_name}" "${link}"

    if [ -h "${prev}" ]
    then
        "${prev}/${server}" stop || true
    fi
fi
rm -f "${app_name}.tar.gz"
"${link}/${server}" start
EOF
}

echo "Deploying..."
scp "${app_file}" "${host}:${deploydir}/${app_name}.tar.gz"
server_commands | ssh "${host}" "bash"

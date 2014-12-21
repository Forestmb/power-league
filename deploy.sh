#!/bin/bash
set -e

cd "$(dirname "$(readlink -f "$0")")"
prog="$(basename "$0")"
function usage()
{
    cat 1>&2 << EOF
Usage: ${prog} [ options ] [ host ]

Description: Package and deploy a instance of the application to a remote host.

Options:
    -B
        Don't build the application before packaging if the binary already exists. If
        this is specified and the binary does not exist, the script will exit with a
        non-zero exit code.
    -v
        Print extra debug information.
EOF
    exit "${1:-1}"
}

build_option=""
verbose="false"
while getopts ":Bhv" option; do
    case "${option}" in
        B)
            build_option="-B"
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

host="$1"
if [ -z "${host}" ]
then
    echo "No host specified." 1>&2
    usage 2
fi

conf="server.conf"
if [ -f "${conf}.${host}" ]
then
    conf="${conf}.${host}"
    log_verbose "Using conf: ${conf}"
fi
. "${conf}"

if [ -z "${deploydir}" ]
then
    deploydir="~/power-league"
fi
log_verbose "Deploy dir: ${deploydir}"

package="./package.sh"

link="current"
prev="previous"
old="old"
server="server.sh"
binary="power-league"
version="$(cat "./.version")"
dist="build/dist"
app="${binary}-${version}-$(date +%Y-%m-%d_%H%M%S)"

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
tar xf "${app}.tar.gz"

mkdir -p "${old}"
if [ -h "${prev}" ]
then
    mv \`readlink -f "${prev}" \` "${old}"
fi

if [ -h "${link}" ]
then
    ln -s -f -T \`readlink -f "${link}" \` "${prev}"
fi
ln -s -f -T "\`pwd\`/${app}" "${link}"

rm -f "${app}.tar.gz"

if [ -h "${prev}" ]
then
    "${prev}/${server}" stop || true
fi
"${link}/${server}" start
EOF
}

# Build package
"${package}" ${build_option} -a "${app}" -c "${conf}"

echo "Deploying..."
scp "${dist}/${app}.tar.gz" "${host}:${deploydir}/${app}.tar.gz"
server_commands | ssh "${host}" "bash"

#!/bin/bash
set -e
cd "$(dirname "$0")"

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
EOF
    exit 1
}

should_build="true"
while getopts ":Bh" option; do
    case "${option}" in
        B)
            should_build="false"
            ;;
        h)
            usage
            ;;
        \?) echo "Unknown option: -${OPTARG}" 1>&2
            usage
            ;;
        *)
            break
            ;;
    esac
done
shift $((OPTIND-1))

host="$1"
conf="server.conf"
if [ -f "${conf}.${host}" ]
then
    conf="${conf}.${host}"
    echo "Using conf: ${conf}"
fi
. "${conf}"

if [ -z "${deploydir}" ]
then
    deploydir="~/power-league"
fi
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

if [ -z "${host}" ]
then
    usage
fi

# Build package

if [ "${should_build}" == "false" ]
then
    "${package}" -a "${app}" -c "${conf}" -B
else
    "${package}" -a "${app}" -c "${conf}"
fi

echo "Deploying..."
cat "${dist}/${app}.tar.gz" | ssh "${host}" "cat > ${deploydir}/${app}.tar.gz"
server_commands | ssh "${host}" "bash"

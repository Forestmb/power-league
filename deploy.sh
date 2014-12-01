#!/bin/bash
set -e
cd "$(dirname "$0")"

prog="$(basename "$0")"
function usage()
{
    echo "Usage: ${prog} host [ deploy-directory ]" 2>&1
    exit 1
}

host="$1"
deploydir="${2:-~/power-league}"
conf="server.conf"
if [ -f "${conf}.${host}" ]
then
    conf="${conf}.${host}"
    echo "Using conf: ${conf}"
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
"${package}" -a "${app}" -c "${conf}"

echo "Deploying..."
cat "${dist}/${app}.tar.gz" | ssh "${host}" "cat > ${deploydir}/${app}.tar.gz"
server_commands | ssh "${host}" "bash"

#!/bin/bash
set -e

cd "$(dirname "$(readlink -f "$0")")"
prog="$(basename "$0")"
function usage
{
    cat 1>&2 << EOF
Usage: ${prog} [ OPTIONS ]

Description: Creates a distributable package to run the power rankings server.

Options:
    -a <app-name>
        Name of the application being packaged. A package of the name
        '<app-name>.tar.gz' will be created.

        Defaults to 'power-league-<version>'. If the version contains '-SNAPSHOT' then
        the current timestamp (YYY-MM-DD_HHMMSS) will be appended to the name.

    -c <conf>
        Server configuration file to use in the package. Defaults to 'server.conf'

    -d <dir>
        Directory to place the packaged files. Defaults to 'build/dist'

    -D <host>
        Deploy the packaged application to a remote host.

    -h
        Display this help.
EOF
    exit "${1:-1}"
}

function snapshot
{
    grep -q -- '-SNAPSHOT' "${1}"
}

binary="power-league"
version="$(cat "./.version")"
dist="build/dist"
deploy_cmd="./deploy.sh"
baseconf="server.conf"

appname="${binary}-${version}"
if snapshot "./.version"
then
    appname="${appname}-$(date +%Y-%m-%d_%H%M%S)"
fi

while getopts ":a:c:d:D:h" option; do
    case "${option}" in
        a)
            appname="${OPTARG}"
            ;;
        c)
            conf="${OPTARG}"
            ;;
        d)
            dist="${OPTARG}"
            ;;
        D)
            host="${OPTARG}"
            ;;
        h)
            usage 0
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

resources=( \
    "LICENSE" \
    "NOTICE" \
    "CHANGELOG.md" \
    ".version" \
    "static" \
    "server.sh" \
    "${binary}" \
)

excluded_resources=( \
    "static/images/originals/" \
)

if [ -z "${conf}" ]
then
    conf="${baseconf}"

    if [ ! -z "${host}" ] && \
       [ -f "${baseconf}.${host}" ]
    then
        conf="${baseconf}.${host}"
    fi
fi

if [ ! -f "${conf}" ]
then
    echo "Error: Server configuration file '${conf}' does not exist." 1>&2
    echo "       Make sure the template file has been copied and edited for this " 1>&2
    echo "       instance of the server." 1>&2
    exit 1
fi

if [ -z "${dist}" ]
then
    echo "Error: '-d' is undefind." 1>&2
    usage 2
fi

if [ -z "${appname}" ]
then
    echo "Error: '-a' is undefined." 1>&2
    usage 3
fi
app="${dist}/${appname}"

if [ ! -f "${binary}" ]
then
    echo "Error: Binary file '${binary}' does not exist, so nothing" 1>&2
    echo "       could be packaged." 1>&2
    exit 4
fi

echo "Packaging..."
if [ -d "${app}" ]
then
    echo "Packaging directory already exists, please remove before continuining" 1>&2
    echo "    dir: ${app}" 1>&2
    exit 5
fi

mkdir -p "${app}"
for resource in "${resources[@]}"; do
    cp -R "${resource}" "${app}"
done
for excluded in "${excluded_resources[@]}"; do
    rm -rf "${app}/${excluded}"
done

# Copy template folder
mkdir "${app}/templates"
cp -R "templates/html" "${app}/templates"

# Copy configuration file separate so it can be renamed
cp "${conf}" "${app}/${baseconf}"

pushd "${dist}" >& /dev/null
tar -zcf "${appname}.tar.gz" "${appname}"
popd >& /dev/null

if [ ! -z "${host}" ]
then
    . "${conf}"
    if [ -z "${deploydir}" ]
    then
        "${deploy_cmd}" "${app}.tar.gz" "${host}"
    else
        "${deploy_cmd}" -d "${deploydir}" "${app}.tar.gz" "${host}"
    fi
fi

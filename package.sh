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
        Name of the application being packaged. Defaults to 'power-rankings-<date>'. A
        package of the name '<app-name>.tar.gz' will be created.
    -B
        Don't build the application before packaging if the binary already exists. If
        this is specified and the binary does not exist, the script will exit with a
        non-zero exit code.
    -c <conf>
        Server configuration file to use in the package. Defaults to 'server.conf'
    -d <dir>
        Directory to place the packaged files. Defaults to 'build/dist'
    -h
        Display this help.
EOF
    exit ${1:-1}
}

binary="power-league"
dist="build/dist"
build_cmd="./build.sh"
should_build="true"
baseconf="server.conf"
conf="${baseconf}"

while getopts ":a:Bc:h" option; do
    case "${option}" in
        a)
            appname="${OPTARG}"
            ;;
        B)
            should_build="false"
            ;;
        c)
            conf="${OPTARG}"
            ;;
        d)
            dist="${OPTARG}"
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

if [ -z "${appname}" ]
then
    appname="${binary}-$(date +%Y-%m-%d_%H%M%S)"
fi
resources=("LICENSE" "NOTICE" "CHANGELOG.md" "static" "server.sh" "${binary}")

if [ ! -f "${conf}" ]
then
    echo "Error: Server configuration file '${conf}' does not exist." 1>&2
    echo "       Make sure the template file has been copied and edited for this " 1>&2
    echo "       instance of the server." 1>&2
    exit 1
fi

app="${dist}/${appname}"

if [ "${should_build}" == "true" ]
then
    "${build_cmd}"
else
    if [ ! -f "${binary}" ]
    then
        echo "Error: No build was requested during packaging but binary file " 1>&2
        echo "       '${binary}' does not exist, so nothing could be packaged." 1>&2
        exit 2
    fi
fi

echo "Packaging..."
rm -rf "${app}"
mkdir -p "${app}"
for resource in ${resources[@]}; do
    cp -R "${resource}" "${app}"
done

# Copy template folder
mkdir "${app}/templates"
cp -R "templates/html" "${app}/templates"

# Copy configuration file separate so it can be renamed
cp "${conf}" "${app}/${baseconf}"

pushd "${dist}" >& /dev/null
tar -zcf "${appname}.tar.gz" "${appname}"
popd >& /dev/null

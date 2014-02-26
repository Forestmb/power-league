#!/bin/bash
set -e

cd "$(dirname "$0")"

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
    -c <conf>
        Server configuration file to use in the package. Defaults to 'server.conf'
    -d <dir>
        Directory to place the packaged files. Defaults to 'build/dist'
    -h
        Display this help.
EOF
    exit 1
}

binary="power-league"
dist="build/dist"
build_cmd="./build.sh"
baseconf="server.conf"
conf="${baseconf}"

while getopts ":a:c:h" option; do
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

if [ -z "${appname}" ]
then
    appname="${binary}-$(date +%Y-%m-%d_%H%M%S)"
fi
resources=("LICENSE" "NOTICE" "static" "server.sh" "${binary}")

if [ ! -f "${conf}" ]
then
    echo "Error: Server configuration file '${conf}' does not exist." 1>&2
    echo "       Make sure the template file has been copied and edited for this " 1>&2
    echo "       instance of the server." 1>&2
    exit 1
fi

app="${dist}/${appname}"

"${build_cmd}"

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

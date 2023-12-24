#!/bin/bash

set -Eeuo pipefail

umask 0022

if [ "$#" -ne 2 ] ; then
    echo "Usage: $0 <BUILDDIR> <SRCDIR>"
    exit 1
fi

# BUILDDIR is the path to the build directory
# SRCDIR is the path to the source directory
#
# e.g. If your current directory is the project root and the build directory is _build, then you want to do
#     $ ./packaging/osx/buildpkg.sh _build .
#

set -u

BUILDDIR="$1"
SRCDIR="$2"

PKGROOTDIR="$(mktemp -d -t unixtools-pkg)"

PKGFILE="${APP_NAME}-${APP_VERSION}.pkg"

dos2unix()
{
    tr -d '\r' < $1 > $2
}

BINDIR="${PKGROOTDIR}/${APP_NAME}-${APP_VERSION}/bin"

mkdir -p -m 0755 "${BINDIR}"

for exe in "${BUILDDIR}"/* ; do
    install -m 0755 "${exe}" "${PKGROOTDIR}/${APP_NAME}-${APP_VERSION}/bin"
done

pkgbuild --root "${PKGROOTDIR}" --install-location ./Library/ \
    --identifier "${APP_IDENTIFIER}" --version "${APP_VERSION}" "${PKGFILE}"

rm -rf $PKGROOTDIR

#productbuild --distribution "$SRCDIR"/packaging/osx/Distribution.xml --resources "$SRCDIR"/packaging/osx/resources --package-path . "$PKGFILE"

rm FoundationDB-clients.pkg FoundationDB-server.pkg
#!/bin/bash

set -e
set -uo pipefail

IDENTIFIER="io.asscrypto.unixtools"
PRODUCT=unixtools
VERSION="$(git describe --abbrev=6)"
BINS="$(find buildd/ -type f | xargs basename)"
TARGETDIR="$(pwd)/pkgbuilddir/target"
BUILDDIR="$(pwd)/buildd"
INSTALLER="${PRODUCT}-macos-installer-${VERSION}.pkg"

substVars() {
    local l_file l_bins

    l_file="${1}"
    l_bins="$(echo ${BINS} | tr -s ' ' ':')"

    sed -i '' -e "s/__IDENTIFIER__/${IDENTIFIER}/g" "${l_file}"
    sed -i '' -e "s/__VERSION__/${VERSION}/g" "${l_file}"
    sed -i '' -e "s/__PRODUCT__/${PRODUCT}/g" "${l_file}"
    sed -i '' -e "s/__BINS__/${l_bins}/g" "${l_file}"
}

BIN_INSTALLDIR="${TARGETDIR}/darwinpkg/Library/${PRODUCT}/${VERSION}"
mkdir -p "${BIN_INSTALLDIR}"
cp -av "${BUILDDIR}"/. "${BIN_INSTALLDIR}"/
chmod -R 0755 "${BIN_INSTALLDIR}"/

cp -rv contrib/darwin "${TARGETDIR}"/
cp LICENSE "${TARGETDIR}/darwin/Resources/"

substVars "${TARGETDIR}/darwin/Distribution"
substVars "${TARGETDIR}/darwin/scripts/postinstall"
substVars "${TARGETDIR}/darwin/Resources/uninstall.sh"

chmod -R 0755 "${TARGETDIR}/darwin/Resources/"
chmod -R 0755 "${TARGETDIR}/darwin/scripts/"
chmod -R 0644 "${TARGETDIR}/darwin/Distribution"

mkdir -p "${TARGETDIR}/package"
chmod 0755 "${TARGETDIR}/package"

mkdir -p "${TARGETDIR}/pkg"

pkgbuild --identifier "${IDENTIFIER}" \
    --version "${VERSION}" \
    --scripts "${TARGETDIR}/darwin/scripts" \
    --root "${TARGETDIR}/darwinpkg" \
    "${TARGETDIR}/package/${PRODUCT}.pkg"

productbuild --distribution "${TARGETDIR}/darwin/Distribution" \
    --resources "${TARGETDIR}/darwin/Resources" \
    --package-path "${TARGETDIR}/package" \
    "${TARGETDIR}/pkg/${INSTALLER}"

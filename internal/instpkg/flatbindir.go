package instpkg

const FlatBinDirUninstall = `#!/bin/bash

DATE=$(date +%Y-%m-%d)
TIME=$(date +%H:%M:%S)
LOG_PREFIX="[$DATE $TIME]"
PRODUCT="{{ (.Name) }}"
VERSION="{{ (.Version) }}"
IDENTIFIER="{{ (.Identifier) }}"
INSTALLDIR="Library/${PRODUCT}/${VERSION}"

#Check running user
if (( $EUID != 0 )); then
    if [ -d "$HOME/${INSTALLDIR}" ]; then
        INSTALLDIR="${HOME}/${INSTALLDIR}"
    else
        echo "Please run as root." >&2
        exit
    fi
else
    INSTALLDIR="/${INSTALLDIR}"
fi

echo "Welcome to Application Uninstaller"
echo "The following packages will be REMOVED:"
echo "  ${PRODUCT} - ${VERSION}"
while true; do
    read -p "Do you wish to continue [Y/n]?" answer
    [[ $answer == "y" || $answer == "Y" || $answer == "" ]] && break
    [[ $answer == "n" || $answer == "N" ]] && exit 0
    echo "Please answer with 'y' or 'n'"
done

#forget from pkgutil
pkgutil --forget "${IDENTIFIER}" 1>&2
if [ $? -eq 0 ]
then
  echo "[2/3] [DONE] Successfully deleted application informations" >&2
else
  echo "[2/3] [ERROR] Could not delete application informations" >&2
fi

#remove application source distribution
[ -e "${INSTALLDIR}" ] && rm -rvf "${INSTALLDIR}"
if [ $? -eq 0 ]
then
  echo "[3/3] [DONE] Successfully deleted application" >&2
else
  echo "[3/3] [ERROR] Could not delete application" >&2
fi

echo "Application uninstall process finished" >&2
exit 0
`

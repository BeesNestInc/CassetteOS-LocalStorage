#!/bin/bash

set -e

readonly CASSETTE_SERVICES=(
    "cassetteos-local-storage.service"
)

readonly CASSETTE_EXEC=cassetteos-local-storage
readonly CASSETTE_CONF=/etc/cassetteos/local-storage.conf
readonly CASSETTE_DB=/var/lib/cassetteos/db/local-storage.db

readonly aCOLOUR=(
    '\e[38;5;154m' # green  	| Lines, bullets and separators
    '\e[1m'        # Bold white	| Main descriptions
    '\e[90m'       # Grey		| Credits
    '\e[91m'       # Red		| Update notifications Alert
    '\e[33m'       # Yellow		| Emphasis
)

Show() {
    # OK
    if (($1 == 0)); then
        echo -e "${aCOLOUR[2]}[$COLOUR_RESET${aCOLOUR[0]}  OK  $COLOUR_RESET${aCOLOUR[2]}]$COLOUR_RESET $2"
    # FAILED
    elif (($1 == 1)); then
        echo -e "${aCOLOUR[2]}[$COLOUR_RESET${aCOLOUR[3]}FAILED$COLOUR_RESET${aCOLOUR[2]}]$COLOUR_RESET $2"
    # INFO
    elif (($1 == 2)); then
        echo -e "${aCOLOUR[2]}[$COLOUR_RESET${aCOLOUR[0]} INFO $COLOUR_RESET${aCOLOUR[2]}]$COLOUR_RESET $2"
    # NOTICE
    elif (($1 == 3)); then
        echo -e "${aCOLOUR[2]}[$COLOUR_RESET${aCOLOUR[4]}NOTICE$COLOUR_RESET${aCOLOUR[2]}]$COLOUR_RESET $2"
    fi
}

Warn() {
    echo -e "${aCOLOUR[3]}$1$COLOUR_RESET"
}

trap 'onCtrlC' INT
onCtrlC() {
    echo -e "${COLOUR_RESET}"
    exit 1
}

if [[ ! -x "$(command -v ${CASSETTE_EXEC})" ]]; then
    Show 2 "${CASSETTE_EXEC} is not detected, exit the script."
    exit 1
fi

while true; do
    echo -n -e "         ${aCOLOUR[4]}Do you want delete local storage database? Y/n :${COLOUR_RESET}"
    read -r input
    case $input in
    [yY][eE][sS] | [yY])
        REMOVE_LOCAL_STORAGE_DATABASE=true
        break
        ;;
    [nN][oO] | [nN])
        REMOVE_LOCAL_STORAGE_DATABASE=false
        break
        ;;
    *)
        Warn "         Invalid input..."
        ;;
    esac
done

for SERVICE in "${CASSETTE_SERVICES[@]}"; do
    Show 2 "Stopping ${SERVICE}..."
    systemctl disable --now "${SERVICE}" || Show 3 "Failed to disable ${SERVICE}"
done

rm -rvf "$(which ${CASSETTE_EXEC})" || Show 3 "Failed to remove ${CASSETTE_EXEC}"
rm -rvf "${CASSETTE_CONF}" || Show 3 "Failed to remove ${CASSETTE_CONF}"

if [[ ${REMOVE_LOCAL_STORAGE_DATABASE} == true ]]; then
    rm -rvf "${CASSETTE_DB}" || Show 3 "Failed to remove ${CASSETTE_DB}"
fi

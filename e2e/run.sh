#!/usr/bin/env bash

# Without ffmpeg, ttyd and jq in binary path, vhs exits with non-zero code, without displaying error message.
ffmpeg -version >/dev/null 2>&1
if [ "$?" -ne 0 ]; then
    echo "Error: 'ffmpeg' is not installed or not found in path."
    echo "Run 'apt install ffmpeg' or 'brew install ffmpeg'."
    exit 1
fi

ttyd -version >/dev/null 2>&1
if [ "$?" -ne 0 ]; then
    echo "Error: 'ttyd' is not installed or not found in path."
    echo "Install ttyd from https://github.com/tsl0922/ttyd, for mac: 'brew install ttyd', for linux: take the binary and put it in your PATH."
    exit 1
fi

jq -V >/dev/null 2>&1
if [ "$?" -ne 0 ]; then
    echo "Error: 'jq' is not installed or not found in path."
    echo "Run apt install jq or brew install jq to install it."
    exit 1
fi

vhs -v >/dev/null 2>&1
if [ "$?" -ne 0 ]; then
    echo "Error: 'vhs' is not installed or not found in path."
    echo "Take the binary from 'https://github.com/charmbracelet/vhs/releases' page."
    exit 1
fi

# cd to this file location
cd "$(dirname "$(readlink -f "$0")")" || exit 1

export VHS_PUBLISH=false # disable "Host your GIF on vhs.charm.sh: vhs publish <file>.gif" message
TMP_HOME=temp
RED_COLOR="\e[31m"
GREEN_COLOR="\e[32m"
NO_COLOR="\e[0m"
MSG_OK="${GREEN_COLOR}OK${NO_COLOR}"
MSG_FAIL="${RED_COLOR}FAIL${NO_COLOR}"

function before_each() {
    local filename_without_extension="$1"
    mkdir -p "${TMP_HOME}"

    # If ssh_config file exists, copy it to the expected directory.
    if [ -f "${filename_without_extension}_ssh_config.yaml" ]; then
        cp "${filename_without_extension}_ssh_config.yaml" "${TMP_HOME}"/ssh_config
    fi

    # If state file exists, copy it to the expected directory.
    if [ -f "${filename_without_extension}_hosts.yaml" ]; then
        cp "${filename_without_extension}_hosts.yaml" "${TMP_HOME}"/hosts.yaml
    fi
}

function check_expected() {
    local filename_without_extension="$1"
    local hosts_file="${TMP_HOME}"/hosts.yaml
    local state_file="${TMP_HOME}"/state.yaml

    # state file should always exist, so we check it first.
    local state_file_expected="expected/${filename_without_extension}_state.yaml"
    diff "${state_file}" "${state_file_expected}"
    if [ "$?" -eq 0 ]; then
        printf "${MSG_OK} %s\n" "${state_file_expected}"
    else
        printf "${MSG_FAIL} %s\n" "${state_file_expected}"
        return 1
    fi

    # Check if hosts file exists in "expected" folder, the run diff.
    if [ -f "${hosts_file}" ]; then
        local hosts_file_expected="expected/${filename_without_extension}_hosts.yaml"
        diff "${hosts_file}" "${hosts_file_expected}"
        if [ "$?" -eq 0 ]; then
            printf "${MSG_OK} %s\n" "${hosts_file_expected}"
        else
            printf "${MSG_FAIL} %s\n" "${hosts_file_expected}"
            return 1
        fi
    fi

    return 0
}

function check_log_no_errors() {
    # Make sure there are no errors in the application log.
    local test_name="$1"
    local app_log="${TMP_HOME}"/app.log
    if grep -qi "ERRO" "${app_log}"; then
        # "ERRO" is not a typo! It's a valid log level name.
        printf "${MSG_FAIL} %s\n" "${app_log}"
        grep -i "ERRO" "${app_log}"
        return 1
    else
        printf "${MSG_OK} %s: %s\n" "${test_name}" "${app_log}"
    fi

    return 0
}

function cleanup() {
    rm -f out.gif "${TMP_HOME}"/*
}

for file in *.tape; do
    cleanup

    filename_without_extension="${file%.*}" # "${file%.*}" = file extension removed
    before_each "$filename_without_extension"

    vhs "$file" >/dev/null 2>&1

    check_expected "$filename_without_extension"
    if [ "$?" -ne 0 ]; then
        printf "${MSG_FAIL} %s\n" "${file}"
        exit 1
    fi

    check_log_no_errors "$filename_without_extension"
    if [ "$?" -ne 0 ]; then
        printf "${MSG_FAIL} %s\n" "Errors found in log file"
        exit 1
    fi

    cleanup
done

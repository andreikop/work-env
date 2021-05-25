#!/bin/bash

function check_if_set {
    var_name=$1
    var_value=${!var_name}
    if [[ -z "$var_value" ]];
    then
        echo $var_name is not set;
        exit -1;
    fi
}

check_if_set WORK_ENV_USER_SHELL
check_if_set WORK_ENV_USER_NAME
check_if_set WORK_ENV_USER_PASSWORD
check_if_set WORK_ENV_USER_ID


if ! id "$WORK_ENV_USER_NAME" &>/dev/null; then
    useradd \
        --shell ${WORK_ENV_USER_SHELL} \
        --groups sudo \
        --uid ${WORK_ENV_USER_ID}  \
        --password "$(openssl passwd -1 ${WORK_ENV_USER_PASSWORD})" \
        ${WORK_ENV_USER_NAME}
fi

user_name=$WORK_ENV_USER_NAME

echo "Welcome to work-env. Your sudo password is '${WORK_ENV_USER_NAME}'"

# Do not pass this variables to users shell
unset WORK_ENV_USER_SHELL
unset WORK_ENV_USER_NAME
unset WORK_ENV_USER_PASSWORD
unset WORK_ENV_USER_ID

exec su ${user_name}

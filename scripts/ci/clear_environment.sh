#!/usr/bin/env bash
# This script is used to clear the environment variables that are set by the
# .env_1011 and .env_local files.  This is useful when switching between
# environments.
# Usage: source clear_environment.sh
# Note: This script must be sourced, not executed.
unset KEYFACTOR_HOSTNAME
unset KEYFACTOR_USERNAME
unset KEYFACTOR_PASSWORD
unset KEYFACTOR_DOMAIN
unset KFUTIL_DEBUG
unset KFUTIL_PROFILE
unset KFUTIL_AUTH_PROVIDER_TYPE
unset KUTIL_AUTH_PROVIDER_PROFILE
unset KUTIL_AUTH_PROVIDER_PARAMS
rm -f "${HOME}/.keyfactor/command_config.json" || true
rm -f output.* || true

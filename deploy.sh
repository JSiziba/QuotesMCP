#!/bin/bash

set -a
source "$(dirname "$0")/.env"
set +a

scp quotes-mcp "$DEPLOY_TARGET"

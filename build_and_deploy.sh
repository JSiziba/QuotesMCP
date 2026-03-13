#!/bin/bash

set -e

DIR="$(dirname "$0")"

"$DIR/build.sh" && "$DIR/deploy.sh"

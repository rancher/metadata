#!/bin/bash
set -x

ip addr add 169.254.169.250/32 dev eth0
echo Starting Metadata Server
exec "$@"

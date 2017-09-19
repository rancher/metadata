#!/bin/bash
set -x -e

if ! ip addr show dev lo | grep -q '169.254.169.250'; then
    ip addr add 169.254.169.250/32 dev lo
fi

echo Adding iptables rules
RANCHER_METADATA_LISTEN_PORT=${RANCHER_METADATA_LISTEN_PORT:-9346}

if ! iptables -t nat -n -L CATTLE_PREROUTING &> /dev/null; then
    iptables -t nat -N CATTLE_PREROUTING
fi


if ! iptables -t nat -n -L CATTLE_OUTPUT &> /dev/null; then
    iptables -t nat -N CATTLE_OUTPUT
fi

if ! iptables -t nat -C CATTLE_PREROUTING -d 169.254.169.250/32 -p tcp -m tcp --dport 80 -j DNAT --to-destination 169.254.169.250:${RANCHER_METADATA_LISTEN_PORT} &> /dev/null; then
    iptables -t nat -A CATTLE_PREROUTING -d 169.254.169.250/32 -p tcp -m tcp --dport 80 -j DNAT --to-destination 169.254.169.250:${RANCHER_METADATA_LISTEN_PORT}
fi

if ! iptables -t nat -C CATTLE_OUTPUT -d 169.254.169.250/32 -p tcp -m tcp --dport 80 -j DNAT --to-destination 169.254.169.250:${RANCHER_METADATA_LISTEN_PORT} &> /dev/null; then
    iptables -t nat -A CATTLE_OUTPUT -d 169.254.169.250/32 -p tcp -m tcp --dport 80 -j DNAT --to-destination 169.254.169.250:${RANCHER_METADATA_LISTEN_PORT}
fi


if ! iptables -t nat -C PREROUTING -m addrtype --dst-type LOCAL -j CATTLE_PREROUTING &> /dev/null; then
    iptables -t nat -I PREROUTING -m addrtype --dst-type LOCAL -j CATTLE_PREROUTING
fi

if ! iptables -t nat -C OUTPUT -m addrtype --dst-type LOCAL -j CATTLE_OUTPUT &> /dev/null; then
    iptables -t nat -I OUTPUT -m addrtype --dst-type LOCAL -j CATTLE_OUTPUT
fi

echo Starting Metadata Server
exec "$@"

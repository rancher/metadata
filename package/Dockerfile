FROM ubuntu:16.04
RUN apt-get update && \
    apt-get install -y iproute2 iptables
COPY metadata start.sh /usr/bin/
CMD ["start.sh", "metadata", "--debug", "--k8s-proxy", "--k8s-proxy-https-listen", ""]

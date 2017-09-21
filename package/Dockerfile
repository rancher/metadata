FROM ubuntu:16.04
RUN apt-get update && \
    apt-get install -y iproute2 iptables
COPY metadata start.sh /usr/bin/
# This is quite obviously insecure and will be removed
COPY ssl /etc/kubernetes/ssl/
CMD ["start.sh", "metadata", "--debug", "--k8s-proxy"]

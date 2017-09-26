FROM rancher/agent-base:v0.3.0
RUN apt-get update && \
    apt-get install -y iproute2 iptables
COPY metadata start.sh /usr/bin/
# This is quite obviously insecure and will be removed
COPY ssl /etc/kubernetes/ssl/

CMD ["start.sh", "metadata", "--debug", "--k8s-proxy"]

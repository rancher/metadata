FROM ubuntu:16.04
RUN apt-get update && \
    apt-get install -y iproute2
COPY metadata start.sh /usr/bin/
CMD ["start.sh", "metadata", "--debug"]

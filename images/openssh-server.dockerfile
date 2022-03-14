FROM ubuntu:20.04
RUN apt-get update \
    && apt-get install --no-install-recommends --yes \
        net-tools openssh-server \
    && rm -rf /var/lib/apt/lists/*
RUN mkdir /run/sshd
CMD ["/usr/sbin/sshd", "-D", "-e"]

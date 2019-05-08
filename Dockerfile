FROM ubuntu

RUN apt-get update && \
    apt-get install -y \
    git \
    golang-go \
    vim
RUN mkdir -p /var/RelayServer/Data
WORKDIR /var/RelayServer/Data
RUN git config core.sparseCheckout true && \
    git remote add -f origin https://methanduck.iptime.org:1300/methanduck/GO && \
    echo "RelaySVR/" >> .git/info/sparse-checkout && \
    git pull origin master
COPY RunServer.sh ./RunServer.sh
ENTRYPOINT [ "RunServer.sh" ]

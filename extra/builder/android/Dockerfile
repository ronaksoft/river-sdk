FROM registry.ronaksoft.com/base/docker/golang
MAINTAINER Ehsan N. Moosa <ehsan@ronaksoft.com>


RUN  go get golang.org/x/mobile/cmd/gomobile
RUN  wget https://github.com/upx/upx/releases/download/v3.96/upx-3.96-amd64_linux.tar.xz
RUN tar -xf upx-3.96-amd64_linux.tar.xz

ENTRYPOINT ["sleep"]
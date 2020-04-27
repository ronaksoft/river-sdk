FROM registry.ronaksoftware.com/base/docker/ubuntu
MAINTAINER Kayvan N. Moosa <kayvannm@ronaksoft.com>

# Add contents from root directory to the docker
ADD ./cli-livelog/ /ronak/bin/cli-livelog

# Set working directory
WORKDIR /ronak/bin

# Entry point
ENTRYPOINT ["/ronak/bin/cli-livelog"]
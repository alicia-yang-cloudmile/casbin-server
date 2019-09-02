FROM grpc/go AS build-server
RUN rm /bin/sh && ln -s /bin/bash /bin/sh

# Install gvm
RUN apt-get -y install bison
RUN curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer > gvm-installer
RUN chmod 755 gvm-installer
RUN ./gvm-installer
RUN rm ./gvm-installer

# # Build code
ADD . /go/src/github.com/casbin/casbin-server
WORKDIR $GOPATH/src/github.com/casbin/casbin-server
RUN protoc -I proto --go_out=plugins=grpc:proto proto/casbin.proto
# Use go1.12
RUN source /root/.gvm/scripts/gvm; gvm install go1.12; gvm use go1.12; go install .; mv $GOPATH/bin/casbin-server /go/bin/casbin-server

FROM grpc/go
COPY --from=build-server go/bin/casbin-server /usr/bin/casbin-server
ENTRYPOINT /usr/bin/casbin-server
EXPOSE 50051

# Install app
# RUN go install .
# ENTRYPOINT $GOPATH/bin/casbin-server

# EXPOSE 50051

FROM golang:1.17 as build
RUN mkdir /package
COPY /internal /package/internal
COPY /sdk /package/sdk
COPY /go.mod /package/go.mod
COPY /go.sum /package/go.sum
WORKDIR /package
RUN go build -v ./...

FROM build as test
COPY /test /package/test
ARG CONDUCTOR_SERVER_URL
ENV CONDUCTOR_SERVER_URL=${CONDUCTOR_SERVER_URL}
RUN go test -v ./...

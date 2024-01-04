ARG GO_VERSION

FROM golang:$GO_VERSION as service_test
ARG SERVICE
RUN echo testing $SERVICE

FROM golang:$GO_VERSION as lib_test
ARG LIB
WORKDIR /usr/src/lib
COPY go.mod ./
RUN go mod download && go mod verify
COPY . .
RUN go test -v ./$LIB/...

FROM golang:$GO_VERSION as controller_build
RUN echo building ${SERVICE}

FROM nginx:1.25.3 AS controller
WORKDIR /app
COPY ./entrypoint /entrypoint
CMD hello world

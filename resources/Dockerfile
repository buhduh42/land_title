ARG GO_VERSION
ARG SERVICE_IMAGE

FROM golang:$GO_VERSION as service_test
ARG SERVICE
RUN echo testing $SERVICE

FROM golang:$GO_VERSION as lib_test
ARG LIB
WORKDIR /usr/src/lib
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go test -v ./$LIB/...

FROM golang:$GO_VERSION as service_build
ARG SERVICE
WORKDIR /usr/src/app
COPY src/go.mod src/
RUN cd src && go mod download && go mod verify
COPY . .
COPY build/entrypoint /entrypoint
RUN cd src && go build -v -o /${SERVICE} .

FROM nginx:1.25.3 AS controller
WORKDIR /app
COPY ./entrypoint /entrypoint
CMD hello world

FROM ${SERVICE_IMAGE} as service
ARG SERVICE
COPY --from=service_build /entrypoint /entrypoint
COPY --from=service_build /$SERVICE /$SERVICE
ENTRYPOINT /entrypoint
CMD /${SERVICE}

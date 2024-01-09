#I want these to fail when not set, but not fail when
#they aren't needed
ARG GO_VERSION=who_cares
ARG SERVICE_IMAGE=who_cares

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

FROM ${SERVICE_IMAGE} as service
ARG SERVICE
ENV SERVICE=$SERVICE
COPY --from=service_build /entrypoint /entrypoint
COPY --from=service_build /$SERVICE /service/
ENTRYPOINT ["/entrypoint"]
CMD /service/$SERVICE

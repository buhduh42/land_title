ARG GO_VERSION

FROM golang:$GO_VERSION as controller_test
RUN echo testing controller

FROM golang:$GO_VERSION as controller_build
RUN echo building controller

FROM nginx:1.25.3 AS controller
WORKDIR /app
COPY ./entrypoint /entrypoint
CMD hello world

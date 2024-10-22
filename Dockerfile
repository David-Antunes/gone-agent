FROM golang:1.23 AS build

WORKDIR /agent

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY internal internal
COPY agent.go .

RUN go build

FROM alpinelinux/docker-cli

COPY --from=build /agent/gone-agent /gone-agent

RUN apk add --no-cache gcompat

CMD ["/gone-agent"]


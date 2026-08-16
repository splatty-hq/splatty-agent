FROM golang:1.26-alpine AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=${VERSION}" -o /out/splatty-agent .

FROM scratch
COPY --from=build /out/splatty-agent /splatty-agent
ENTRYPOINT ["/splatty-agent"]

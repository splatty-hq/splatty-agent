FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/splatty-agent .

FROM scratch
COPY --from=build /out/splatty-agent /splatty-agent
ENTRYPOINT ["/splatty-agent"]

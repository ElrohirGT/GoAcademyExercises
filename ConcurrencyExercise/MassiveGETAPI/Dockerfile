FROM golang:1.24 AS build
WORKDIR /src

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
# COPY go.mod go.sum ./
# RUN go mod download

COPY . /src

WORKDIR /src/MassiveGETAPI
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/goapp .
# RUN timeout 3s /bin/goapp || 0

# FROM debian:latest
FROM alpine:3.14
COPY --from=build /bin/goapp /bin/goapp
# ENTRYPOINT ["tail", "-f", "/dev/null"]
CMD ["/bin/goapp"]

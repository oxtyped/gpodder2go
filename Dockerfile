FROM golang:1.21.4 AS Build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY cmd ./cmd
COPY pkg ./pkg
RUN CGO_ENABLED=0 GOOS=linux go build -o /gpodder2go

FROM alpine:3.18.4
RUN mkdir /data
WORKDIR /data
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /data /entrypoint.sh
COPY --from=Build /gpodder2go /gpodder2go

EXPOSE 3005
VOLUME /data
ENTRYPOINT ["/entrypoint.sh"]

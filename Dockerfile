FROM golang:1.14

WORKDIR /mnt
COPY src /mnt
RUN go build -ldflags "-linkmode external -extldflags -static" -a main.go

FROM scratch
COPY --from=0 /mnt/main /library
COPY --from=0 /etc/ssl/certs/ca-certificates.crt  /etc/ssl/certs/ca-certificates.crt

CMD ["/library", "serve"]
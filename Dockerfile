FROM golang:1.15 AS builder
RUN mkdir /app
WORKDIR /app

# Dependencies
ADD go.mod go.sum /app/
RUN go mod download && go mod verify
RUN GO111MODULE=off go get -u golang.org/x/lint/golint
RUN mkdir certs && \
    printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost,IP:127.0.0.1\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth" | \
    openssl req -x509 -out certs/localhost.crt -keyout certs/localhost.key \
            -days 825 -newkey rsa:2048 -nodes -sha256 -subj '/CN=localhost' -extensions EXT -config -
   
# Source
ADD static /app/static
ADD templates /app/templates
ADD rankings /app/rankings
ADD session /app/session
ADD site /app/site
ADD power-league.go /app
ADD build.sh /app

# Build
RUN ./build.sh

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/certs certs
COPY --from=builder /app/static static
COPY --from=builder /app/templates templates
COPY --from=builder /app/power-league .
ENTRYPOINT [ "./power-league", "-logtostderr" ]

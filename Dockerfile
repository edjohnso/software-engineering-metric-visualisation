# Build
FROM golang:latest AS build
ARG GHO_CLIENT_ID
ARG GHO_CLIENT_SECRET
ARG GHO_PAT
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
COPY pkg pkg
RUN go vet ./pkg/webserver
RUN CGO_ENABLED=0 go build ./cmd/webserver
RUN go test -cover ./pkg/webserver

# Deploy
FROM scratch
ARG GHO_CLIENT_ID
ARG GHO_CLIENT_SECRET
ENV GHO_CLIENT_ID=$GHO_CLIENT_ID GHO_CLIENT_SECRET=$GHO_CLIENT_SECRET
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/webserver .
COPY web web
EXPOSE 80
ENTRYPOINT ["/webserver"]

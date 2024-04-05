FROM golang:1.21 AS build-stage

WORKDIR /term-api

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-bash-api

FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /docker-bash-api /docker-bash-api

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/docker-bash-api"]
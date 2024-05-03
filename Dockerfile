FROM golang:1.21

WORKDIR /term-api

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-bash-api

EXPOSE 8080
# USER nonroot:nonroot
CMD ["/docker-bash-api"]
FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
# RUN go mod download

COPY ./api ./api
COPY ./db ./db
COPY ./db/migrations ./db/migrations

COPY . .

RUN chmod +x /app/main

CMD ["/app/main"]

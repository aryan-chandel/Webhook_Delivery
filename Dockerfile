FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o webhook_service

EXPOSE 8000

CMD [ "./webhook_service" ]
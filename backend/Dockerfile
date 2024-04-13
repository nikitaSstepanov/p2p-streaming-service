FROM golang:alpine AS builder

EXPOSE 80

WORKDIR /usr/local/src

COPY ["go.mod", "go.sum", "./"]

RUN ["go", "mod", "download"]

COPY ./ ./

RUN ["go", "build", "-o", "./bin/p2p-streaming-service", "./cmd/p2p-streaming-service"]

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/p2p-streaming-service ./

COPY ./config/config.yml ./config/config.yml

COPY ./migrations/scheme ./migrations/scheme

COPY ./files ./files

COPY .env .env

CMD [ "./p2p-streaming-service" ]
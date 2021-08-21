FROM node:16-alpine AS frontend-builder

RUN mkdir /app
WORKDIR /app

COPY webui/package.json .
COPY webui/package-lock.json .
RUN npm install

COPY webui .
RUN npx next build
RUN npx next export


###############

FROM golang:1.16 AS server-builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cli cli
COPY cmd cmd
COPY lib lib
COPY server server
RUN CGO_ENABLED=0 GOOS=linux go build -tags production -o sqedule-server -ldflags '-w -s' -a -installsuffix cgo ./cmd/server


###############

FROM alpine:3.14

LABEL maintainer="Fullstaq"

COPY --from=frontend-builder /app/out webui-assets
COPY --from=server-builder /app/sqedule-server .

EXPOSE 3001

# Using entrypoint so we can use commands in docker compose. Rather than using CMD
ENTRYPOINT ["./sqedule-server"]

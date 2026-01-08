FROM golang:1.24 AS builder
WORKDIR /app/backend

# copy go modules from project root (build context is project root)
COPY backend/go.mod backend/go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
RUN go mod download

# copy backend source
COPY backend ./

# copy frontend into the image so final stage can include it
COPY frontend /app/frontend

# Build static binary (build in backend folder)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app

# copy binary and frontend files
COPY --from=builder /server /server
COPY --from=builder /app/frontend /app/frontend

# create non-root user and give ownership to /app so sqlite can create the DB
RUN addgroup -S app && adduser -S app -G app && chown -R app:app /app

EXPOSE 3000
VOLUME ["/app"]

USER app
ENTRYPOINT ["/server"]

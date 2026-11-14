# syntax=docker/dockerfile:1

# ---- Build stage: full Go toolchain ----
FROM golang:1.27-alpine AS build
WORKDIR /src

# Copy dependency files first so this layer is cached
# until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- Runtime stage: binary only ----
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/server /server

ENV GIN_MODE=release \
    HTTP_ADDR=:8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
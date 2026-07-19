# syntax=docker/dockerfile:1
#
# Build context is the repository root (see docker-compose.yml) so the server
# can be bundled with the OpenAPI contract it serves at runtime.

FROM golang:1.25.7 AS build
WORKDIR /src/backend

# Cache module downloads before copying the rest of the source.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# Distroless "static" ships CA certificates (needed for outbound HTTPS activity
# nodes) and runs as a non-root user by default.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /out/server /app/server
COPY api/openapi.yaml /app/api/openapi.yaml

ENV OPENAPI_SPEC_PATH=/app/api/openapi.yaml
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]

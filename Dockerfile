FROM node:24-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run generate

FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/frontend/.output/public ./frontend/.output/public
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" -o /great-firewall ./cmd/gfw

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=backend /great-firewall /app/great-firewall
COPY --from=frontend /src/frontend/.output/public /app/frontend/.output/public
ENV GFW_LISTEN_ADDRESS=:8080 GFW_FRONTEND_DIR=/app/frontend/.output/public
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/great-firewall"]
CMD ["serve"]

# Etapa 1: Build
FROM golang:1.23.4 AS builder

WORKDIR /app

# Copiar los archivos de módulos y descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar todo el código fuente
COPY . .

# Compilar binario estático para Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Etapa 2: Imagen ligera para producción
FROM alpine:3.18

# Instalar certificados SSL para conexiones TLS y Swagger
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copiar el binario y carpetas necesarias desde builder
COPY --from=builder /app/main .
COPY --from=builder /app/uploads ./uploads
COPY --from=builder /app/docs ./docs

# Puerto donde corre la API
EXPOSE 8082

# Comando para arrancar la app
CMD ["./main"]

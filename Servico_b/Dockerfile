# Fase de build
FROM golang:1.23 as build

# Define o diretório de trabalho
WORKDIR /app

# Copia o restante dos arquivos para o contêiner
COPY . .

# Define o diretório específico do comando
WORKDIR /app/cmd

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o 20cloudrun .

# Fase final
FROM golang:1.23-alpine  
WORKDIR /app
COPY --from=build /app/cmd/20cloudrun .
ENTRYPOINT ["./20cloudrun"]
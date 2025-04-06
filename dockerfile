# Базовый образ для сборки
FROM ubuntu:24.04 AS builder

# Установка зависимостей
RUN apt update && apt install -y \
    curl \
    sshpass \
    openssh-client \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Ручная установка Go 1.22.10 (стабильная версия)
RUN curl -LO https://go.dev/dl/go1.22.10.linux-amd64.tar.gz && \
    echo "736ce492a19d756a92719a6121226087ccd91b652ed5caec40ad6dbfb2252092  go1.22.10.linux-amd64.tar.gz" | sha256sum -c - && \
    tar -C /usr/local -xzf go1.22.10.linux-amd64.tar.gz && \
    rm go1.22.10.linux-amd64.tar.gz

# Настройка переменных окружения
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV PATH=$PATH:$GOPATH/bin

# Обновление сертификатов
RUN update-ca-certificates

# Настройка рабочей директории
WORKDIR /app

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN go mod download && \
    go mod tidy && \
    go build -o nvidia-vgpu-exporter ./main.go

# Финальный образ
FROM ubuntu:24.04

RUN apt update && apt install -y \
    sshpass \
    openssh-client \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/nvidia-vgpu-exporter /usr/local/bin/
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
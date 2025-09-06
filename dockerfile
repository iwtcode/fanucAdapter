FROM golang:bullseye as builder

WORKDIR /tmp/
# Копируем x86_64 версию библиотеки
COPY libfwlib32.so* fwlib32.h ./

# Устанавливаем библиотеку в систему
RUN cp libfwlib32.so /usr/lib/ && \
    ldconfig

WORKDIR /usr/src/app
COPY . .

# Нативная сборка для x86_64
RUN go build -o fwlib_example .

# Финальный образ
FROM debian:bullseye-slim

WORKDIR /app

# Копируем бинарник
COPY --from=builder /usr/src/app/fwlib_example .

# Копируем библиотеку
COPY --from=builder /tmp/libfwlib32.so /usr/lib/

# Устанавливаем минимальные зависимости
RUN apt-get update && apt-get install -y \
    libc6 \
    && rm -rf /var/lib/apt/lists/* && \
    ldconfig

CMD ["./fwlib_example"]
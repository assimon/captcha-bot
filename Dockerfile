FROM debian:bullseye-slim
RUN apt-get update -y \
    && apt-get install -y libsqlite3-dev \
    && apt-get install -y sqlite3 \
    && apt-get install -y libc6 \
    && apt-get install -y libc6-dev \
    && apt-get install -y musl
RUN mkdir -p /app/db
ADD config.toml /app/
ADD captcha-bot /app/
ADD cacert.pem /etc/ssl/certs/
WORKDIR /app
ENTRYPOINT ["./captcha-bot"]
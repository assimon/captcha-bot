FROM debian:bullseye-slim
RUN apt-get update -y \
    && apt-get install -y libsqlite3-dev \
    && apt-get install -y sqlite3 \
    && apt-get install -y libc6 \
    && apt-get install -y libc6-dev
RUN mkdir /app
ADD cacert.pem /etc/ssl/certs/
WORKDIR /app
ENTRYPOINT ["/captcha-bot"]
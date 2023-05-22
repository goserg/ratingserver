FROM ubuntu
ENV TZ="Europe/Moscow"

LABEL owner="goserg" \
      maintainer="goserg"

WORKDIR /app/server

COPY bin/server /app/server
COPY cert.pem /app/server
COPY key.pem /app/server

EXPOSE 9865

CMD ["/app/server/server"]
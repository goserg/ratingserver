FROM ubuntu
ENV TZ="Europe/Moscow"

LABEL owner="goserg" \
      maintainer="goserg"

WORKDIR /app/server

COPY bin/server /app/server

EXPOSE 3000

CMD ["/app/server/server"]
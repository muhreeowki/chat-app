FROM golang:1.23.0-bullseye

WORKDIR /app 

COPY . .

RUN go build -o ./chatserver

EXPOSE 4000

EXPOSE 8080

ENV JWT_SECRET="LivingInUnionWithChrist"

ENV DB_CONN_STR="postgres://chatadmin:chatapppassword@chatdb/chatdb?sslmode=disable"

CMD [ "./chatserver" ]

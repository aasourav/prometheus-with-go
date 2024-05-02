FROM golang

WORKDIR /app

COPY . .

ENTRYPOINT [ "./prometheus" ]
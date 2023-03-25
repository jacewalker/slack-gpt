FROM golang:1.20.2
WORKDIR /app
COPY . /app/
WORKDIR /app/
RUN go mod download
RUN go build -buildvcs=false -o main .
EXPOSE 80
EXPOSE 443
CMD ["/app/main"]

# docker build -t jaceau:latest . --no-cache
# docker run --name jace.au -itd --network=proxy -p :8080 jaceau:latest
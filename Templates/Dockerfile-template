# Build binary
FROM golang:alpine as build

WORKDIR /build

COPY go.mod go.sum ./
COPY . .

RUN GOOS=linux go build -o ./main main.go

# Start lambda container from fresh image 
FROM scratch


WORKDIR "/var/task"
COPY --from=build /build/main /var/task

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.7.0 /lambda-adapter /opt/extensions/lambda-adapter




CMD ["/var/task/main"]
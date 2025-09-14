# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY get_station_details_lambda/ ./get_station_details_lambda/
COPY scrapper/ ./scrapper/

WORKDIR /app/get_station_details_lambda
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go init.go

# Runtime stage
FROM public.ecr.aws/lambda/provided:al2-x86_64

COPY --from=builder /app/get_station_details_lambda/bootstrap ${LAMBDA_RUNTIME_DIR}/

CMD ["bootstrap"]
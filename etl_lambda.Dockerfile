# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY etl_lambda/ ./etl_lambda/
COPY scrapper/ ./scrapper/

WORKDIR /app/etl_lambda
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go init.go

# Runtime stage
FROM public.ecr.aws/lambda/provided:al2-x86_64

COPY --from=builder /app/etl_lambda/bootstrap ${LAMBDA_RUNTIME_DIR}

CMD ["bootstrap"]
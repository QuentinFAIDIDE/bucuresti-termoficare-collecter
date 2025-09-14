# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY get_counts_lambda/ ./get_counts_lambda/
COPY scrapper/ ./scrapper/

WORKDIR /app/get_counts_lambda
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go init.go

# Runtime stage
FROM public.ecr.aws/lambda/provided:al2-x86_64

COPY --from=builder /app/get_counts_lambda/bootstrap ${LAMBDA_RUNTIME_DIR}/

CMD ["bootstrap"]
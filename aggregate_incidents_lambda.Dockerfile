# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY aggregate_incidents_lambda/ ./aggregate_incidents_lambda/
COPY scrapper/ ./scrapper/

WORKDIR /app/aggregate_incidents_lambda
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go init.go

# Runtime stage
FROM public.ecr.aws/lambda/provided:al2-x86_64

COPY --from=builder /app/aggregate_incidents_lambda/bootstrap ${LAMBDA_RUNTIME_DIR}

CMD ["bootstrap"]
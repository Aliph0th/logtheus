# Logtheus

Logtheus is a microservice-based platform for ingesting, processing, and storing application logs.
It combines Go services, a Python ingestion workers with ML model, gRPC APIs, and Kafka-based asynchronous pipelines.

## Architecture Overview

- API Gateway handles external HTTP traffic and routes requests to internal services.
- Core domain services (User, Project, Application) manage identity, ownership, and API keys.
- Ingestion service receives raw logs, validates source API keys, normalizes payloads, and asynchronously processes them through a background worker pool. ML-based attribute extraction can enrich logs with structured fields before Kafka publishing.
- Log Engine consumes log batches from Kafka and persists data to storage.
- Mail service consumes Kafka events and sends verification/invite emails asynchronously.

## Services

- `gateway`: Public HTTP entrypoint.
- `user-service`: User accounts, tokens, and auth-related operations.
- `project-service`: Projects, members, invitations.
- `application-service`: Application metadata and API key validation.
- `ingestion-service`: Python gRPC ingestion service and parsing pipeline.
  - Receives raw logs over gRPC.
  - Background worker pool (configurable count + queue size) processes batches asynchronously.
  - ML-based attribute extraction: uses a token classification (NER) self-trained model from Hugging Face, extracts structured fields (service, level, timestamp, error_code, etc.), and includes confidence scores.
  - Returns 202 Accepted immediately; processing happens in the background.
- `log-engine-service`: Kafka consumer + storage writer for logs.
- `mail-service`: Email worker consuming Kafka topics.
- `shared`: Common protobuf contracts, clients, constants, and utility code.

## Infrastructure

Managed in Docker Compose:

- PostgreSQL (user DB)
- PostgreSQL (project DB)
- Redis
- ClickHouse
- Kafka (KRaft)

## Messaging

Kafka is used for asynchronous communication:

- Mail events (`mail.verify`, `mail.invite`)
- Log ingestion batches (`logs.ingest.v1`)

This decouples request handling from downstream processing and improves burst tolerance.

## Getting Started

### 1. Start infrastructure

Run from `docker` folder:

```bash
docker compose up -d
```

### 2. Generate protobuf code

```bash
task protoc
task protoc:python
```

### 3. Run services in development mode

```bash
task dev:all
```

## ML Model Setup (Ingestion Service)

The ingestion service can extract structured attributes from unformatted logs using a PyTorch-based token classification model.
[my github repo](https://github.com/Aliph0th/logtheus-ml)

### Configuration

Set these environment variables in `ingestion-service/.env`:

- `LOG_MODEL_HF_SOURCE`: Hugging Face repo ID or full URL (uses [logtheus-ml-base](https://huggingface.co/Aliph0th/logtheus-ml-base)) (but there's a large version [logtheus-ml-large](https://huggingface.co/Aliph0th/logtheus-ml-large))
  - Leave empty to disable ML extraction.
- `LOG_MODEL_LOCAL_DIR`: Local directory where downloaded model will be cached
- `LOG_MODEL_REVISION`: Model revision/branch on Hugging Face
- `LOG_MODEL_HF_TOKEN`: HF authentication token if model is private (optional)
- `LOG_MODEL_CONFIDENCE_THRESHOLD`: Minimum confidence score for extracted attributes
- `INGEST_WORKER_COUNT`: Number of background workers processing logs (default: `2`)
- `INGEST_QUEUE_SIZE`: Max queued jobs before rejecting requests (default: `2000`)

### How It Works

1. On service startup, the model is downloaded from Hugging Face (if not already cached) and loaded into memory.
2. When logs arrive via gRPC, they are queued immediately and the client receives a 202 Accepted response.
3. Background workers extract structured attributes (level, service, user_id, error_code, etc.) and compute confidence scores using the loaded ML model.
4. Logs with extracted attributes plus `ml_confidence` and `ml_low_confidence_attributes` are published to Kafka.
5. Downstream processors (log-engine-service) save enriched logs to storage.

### Output Format

Extracted attributes appear in the `attributes` map:

```json
{
  "service": "auth",
  "level": "error",
  "ml_confidence": "0.891234"
}
```

Low-confidence fields are stored separately in `ml_low_confidence_attributes` for later review or model retraining.

## Notes

- The Python ingestion service uses its own virtual environment in `ingestion-service/.venv`.
- Configuration is environment-driven via each service `.env` file.
- gRPC protobuf files under `shared/proto` are the source of truth for service contracts.
- Background worker threads in ingestion-service handle all ML inference and Kafka publishing, keeping the RPC path fast.
- Model artifacts (tokenizer, weights, config) must be in safetensors or PyTorch checkpoint format.
- The service gracefully drains the job queue on shutdown, ensuring no logs are dropped.

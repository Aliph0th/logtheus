from dataclasses import dataclass
import os
from dotenv import load_dotenv


@dataclass
class AppConfig:
    port: int
    env: str
    application_service: str
    kafka_brokers: str
    kafka_topic: str
    kafka_features_topic: str
    kafka_username: str
    kafka_password: str
    kafka_mechanism: str
    kafka_api_version: str
    model_hf_source: str
    model_local_dir: str
    model_revision: str
    model_hf_token: str
    model_confidence_threshold: float
    ingest_worker_count: int
    ingest_queue_size: int


def load_config() -> AppConfig:
    load_dotenv()

    return AppConfig(
        port=int(os.getenv("PORT", "50057")),
        env=os.getenv("ENV", "development"),
        application_service=os.getenv(
            "APPLICATION_SERVICE", "localhost:50054"),
        kafka_brokers=os.getenv("KAFKA_BROKERS", "localhost:9094"),
        kafka_topic=os.getenv("KAFKA_TOPIC", "logs.ingest.v1"),
        kafka_features_topic=os.getenv("KAFKA_FEATURES_TOPIC", "logs.features.v1"),
        kafka_username=os.getenv("KAFKA_USERNAME", ""),
        kafka_password=os.getenv("KAFKA_PASSWORD", ""),
        kafka_mechanism=os.getenv("KAFKA_MECHANISM", "PLAIN"),
        kafka_api_version=os.getenv("KAFKA_API_VERSION", "3.7.0"),
        model_hf_source=os.getenv("LOG_MODEL_HF_SOURCE", "").strip(),
        model_local_dir=os.getenv(
            "LOG_MODEL_LOCAL_DIR", "./models/log-attribute-model"),
        model_revision=os.getenv("LOG_MODEL_REVISION", "main"),
        model_hf_token=os.getenv("LOG_MODEL_HF_TOKEN", "").strip(),
        model_confidence_threshold=float(
            os.getenv("LOG_MODEL_CONFIDENCE_THRESHOLD", "0.85")),
        ingest_worker_count=int(os.getenv("INGEST_WORKER_COUNT", "2")),
        ingest_queue_size=int(os.getenv("INGEST_QUEUE_SIZE", "2000")),
    )

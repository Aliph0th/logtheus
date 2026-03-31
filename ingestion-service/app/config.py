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
    kafka_username: str
    kafka_password: str
    kafka_mechanism: str
    kafka_api_version: str



def load_config() -> AppConfig:
    load_dotenv()

    return AppConfig(
        port=int(os.getenv("PORT", "50057")),
        env=os.getenv("ENV", "development"),
        application_service=os.getenv("APPLICATION_SERVICE", "localhost:50054"),
        kafka_brokers=os.getenv("KAFKA_BROKERS", "localhost:9094"),
        kafka_topic=os.getenv("KAFKA_TOPIC", "logs.ingest.v1"),
        kafka_username=os.getenv("KAFKA_USERNAME", ""),
        kafka_password=os.getenv("KAFKA_PASSWORD", ""),
        kafka_mechanism=os.getenv("KAFKA_MECHANISM", "PLAIN"),
        kafka_api_version=os.getenv("KAFKA_API_VERSION", "3.7.0"),
    )

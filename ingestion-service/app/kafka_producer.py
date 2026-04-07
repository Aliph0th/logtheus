from kafka import KafkaProducer


class LogsKafkaProducer:
    def __init__(
        self,
        brokers: str,
        topic: str,
        username: str,
        password: str,
        mechanism: str,
        api_version: str,
    ) -> None:
        servers = [item.strip() for item in brokers.split(",") if item.strip()]
        kwargs = {
            "bootstrap_servers": servers,
            "acks": 1,
            "linger_ms": 20,
            "retries": 3,
            "max_block_ms": 5000,
            "api_version": self._parse_api_version(api_version),
        }

        if username and password:
            kwargs["security_protocol"] = "SASL_PLAINTEXT"
            kwargs["sasl_mechanism"] = mechanism or "PLAIN"
            kwargs["sasl_plain_username"] = username
            kwargs["sasl_plain_password"] = password

        self._topic = topic
        self._producer = KafkaProducer(**kwargs)

    @staticmethod
    def _parse_api_version(value: str) -> tuple[int, int, int]:
        parts = [part.strip() for part in value.split(".") if part.strip()]
        if len(parts) != 3:
            return (3, 7, 0)

        try:
            return (int(parts[0]), int(parts[1]), int(parts[2]))
        except ValueError:
            return (3, 7, 0)

    def publish(self, payload: bytes, key: bytes | None = None) -> None:
        future = self._producer.send(self._topic, value=payload, key=key)
        future.get(timeout=5)

    def close(self) -> None:
        self._producer.flush(timeout=5)
        self._producer.close()

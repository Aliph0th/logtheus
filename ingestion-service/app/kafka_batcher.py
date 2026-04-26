from __future__ import annotations

from dataclasses import dataclass
from queue import Empty, Queue
from threading import Event, Thread

from app.proto import logengine_pb2


@dataclass
class _BatchEnvelope:
    logs: list[logengine_pb2.LogItem]
    features: list[logengine_pb2.LogFeatureItem]


class KafkaPublishBatcher:
    def __init__(
        self,
        max_logs: int,
        max_bytes: int,
        max_wait_ms: int,
        publish_logs,
        publish_features,
    ) -> None:
        self._max_logs = max(1, max_logs)
        self._max_bytes = max(1024, max_bytes)
        self._max_wait_seconds = max(1, max_wait_ms) / 1000
        self._publish_logs = publish_logs
        self._publish_features = publish_features

        self._queue: Queue[_BatchEnvelope | None] = Queue()
        self._stop_event = Event()
        self._thread = Thread(target=self._run, name="kafka-batch-publisher", daemon=True)
        self._thread.start()

    def enqueue(self, logs: list[logengine_pb2.LogItem], features: list[logengine_pb2.LogFeatureItem]) -> None:
        if len(logs) == 0:
            return
        self._queue.put(_BatchEnvelope(logs=logs, features=features))

    def close(self) -> None:
        self._stop_event.set()
        self._queue.put(None)
        self._thread.join(timeout=5)

    def _run(self) -> None:
        buffer_logs: list[logengine_pb2.LogItem] = []
        buffer_features: list[logengine_pb2.LogFeatureItem] = []
        buffer_bytes = 0

        while True:
            timeout = self._max_wait_seconds if len(buffer_logs) > 0 else 0.2
            try:
                envelope = self._queue.get(timeout=timeout)
            except Empty:
                if len(buffer_logs) > 0:
                    self._flush(buffer_logs, buffer_features)
                    buffer_logs = []
                    buffer_features = []
                    buffer_bytes = 0
                if self._stop_event.is_set():
                    return
                continue

            if envelope is None:
                if len(buffer_logs) > 0:
                    self._flush(buffer_logs, buffer_features)
                return

            envelope_bytes = sum(len(item.raw_data) for item in envelope.logs)
            next_logs = len(buffer_logs) + len(envelope.logs)
            next_bytes = buffer_bytes + envelope_bytes

            if len(buffer_logs) > 0 and (next_logs > self._max_logs or next_bytes > self._max_bytes):
                self._flush(buffer_logs, buffer_features)
                buffer_logs = []
                buffer_features = []
                buffer_bytes = 0

            buffer_logs.extend(envelope.logs)
            buffer_features.extend(envelope.features)
            buffer_bytes += envelope_bytes

            if len(buffer_logs) >= self._max_logs or buffer_bytes >= self._max_bytes:
                self._flush(buffer_logs, buffer_features)
                buffer_logs = []
                buffer_features = []
                buffer_bytes = 0

    def _flush(self, logs: list[logengine_pb2.LogItem], features: list[logengine_pb2.LogFeatureItem]) -> None:
        logs_payload = logengine_pb2.SaveLogsRequest(logs=logs).SerializeToString()
        project_key = str(logs[0].project_id).encode("utf-8")
        self._publish_logs(logs_payload, key=project_key)

        if len(features) > 0:
            features_payload = logengine_pb2.SaveLogFeaturesRequest(features=features).SerializeToString()
            self._publish_features(features_payload, key=project_key)

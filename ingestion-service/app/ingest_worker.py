from __future__ import annotations

import logging
from dataclasses import dataclass
from queue import Empty, Full, Queue
from threading import Event, Thread
from app.constants import LogFormat
from typing import Callable


class QueueFullError(Exception):
    pass


@dataclass
class PreparedLogEntry:
    raw_data: bytes
    fmt: LogFormat
    attrs: dict[str, str]
    embedding: list[float]


@dataclass
class IngestLogPayload:
    raw_data: bytes


@dataclass
class IngestJob:
    application_id: int
    application_name: str
    project_id: int
    source_ip: str
    logs: list[IngestLogPayload]


class IngestWorkerPool:
    def __init__(
        self,
        worker_count: int,
        queue_size: int,
        processor: Callable[[IngestJob], None],
    ) -> None:
        self._processor = processor
        self._jobs: Queue[IngestJob | None] = Queue(maxsize=max(1, queue_size))
        self._stop_event = Event()
        self._workers: list[Thread] = []

        final_worker_count = max(1, worker_count)
        for index in range(final_worker_count):
            worker = Thread(
                target=self._worker_loop,
                name=f"ingestion-worker-{index + 1}",
                daemon=True,
            )
            worker.start()
            self._workers.append(worker)

    @property
    def worker_count(self) -> int:
        return len(self._workers)

    @property
    def queue_size(self) -> int:
        return self._jobs.maxsize

    def enqueue(self, job: IngestJob) -> None:
        try:
            self._jobs.put_nowait(job)
        except Full as exc:
            raise QueueFullError("Ingestion queue is full") from exc

    def close(self) -> None:
        for _ in self._workers:
            try:
                self._jobs.put(None, timeout=2)
            except Exception:
                logging.warning(
                    "Failed to enqueue sentinel for worker shutdown")
        self._stop_event.set()
        for worker in self._workers:
            worker.join(timeout=5)

    def _worker_loop(self) -> None:
        while not self._stop_event.is_set():
            try:
                job = self._jobs.get(timeout=0.2)
            except Empty:
                continue

            if job is None:
                self._jobs.task_done()
                return

            try:
                self._processor(job)
            except Exception:
                logging.exception("Background ingestion job failed")
            finally:
                self._jobs.task_done()

from __future__ import annotations

import logging
from dataclasses import dataclass
from collections import deque
from threading import Condition, Event, Thread
from app.constants import LogFormat
from typing import Callable


class QueueFullError(Exception):
    pass


class ProjectQueueFullError(QueueFullError):
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
        per_project_queue_limit: int,
        overload_high_watermark: float,
        processor: Callable[[IngestJob], None],
    ) -> None:
        self._processor = processor
        self._queue_size = max(1, queue_size)
        self._per_project_queue_limit = max(1, per_project_queue_limit)
        self._overload_high_watermark = min(max(overload_high_watermark, 0.1), 1.0)

        self._total_jobs = 0
        self._project_jobs: dict[int, int] = {}
        self._project_queues: dict[int, deque[IngestJob]] = {}
        self._active_projects: deque[int] = deque()
        self._is_active_project: set[int] = set()

        self._condition = Condition()
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
        return self._queue_size

    @property
    def queue_length(self) -> int:
        with self._condition:
            return self._total_jobs

    @property
    def queue_fill_ratio(self) -> float:
        with self._condition:
            return self._total_jobs / self._queue_size

    def enqueue(self, job: IngestJob) -> None:
        with self._condition:
            if self._total_jobs >= self._queue_size:
                raise QueueFullError("Ingestion queue is full")

            project_pending = self._project_jobs.get(job.project_id, 0)
            if project_pending >= self._per_project_queue_limit:
                raise ProjectQueueFullError("Project queue is full")

            overload_threshold = int(self._queue_size * self._overload_high_watermark)
            active_projects = max(1, len(self._project_jobs) + (0 if project_pending > 0 else 1))
            reserved_per_project = max(1, overload_threshold // active_projects)
            if self._total_jobs >= overload_threshold and project_pending >= reserved_per_project:
                raise ProjectQueueFullError("Project exceeded overload quota")

            project_queue = self._project_queues.get(job.project_id)
            if project_queue is None:
                project_queue = deque()
                self._project_queues[job.project_id] = project_queue

            project_queue.append(job)
            self._project_jobs[job.project_id] = project_pending + 1
            self._total_jobs += 1

            if job.project_id not in self._is_active_project:
                self._active_projects.append(job.project_id)
                self._is_active_project.add(job.project_id)

            self._condition.notify()

    def close(self) -> None:
        self._stop_event.set()
        with self._condition:
            self._condition.notify_all()
        for worker in self._workers:
            worker.join(timeout=5)

    def _worker_loop(self) -> None:
        while True:
            job: IngestJob | None = None
            with self._condition:
                while self._total_jobs == 0 and not self._stop_event.is_set():
                    self._condition.wait(timeout=0.2)

                if self._total_jobs == 0 and self._stop_event.is_set():
                    return

                job = self._pop_next_job()

            if job is None:
                if self._stop_event.is_set():
                    return
                continue

            try:
                self._processor(job)
            except Exception:
                logging.exception("Background ingestion job failed")

    def _pop_next_job(self) -> IngestJob | None:
        for _ in range(len(self._active_projects)):
            project_id = self._active_projects.popleft()
            queue = self._project_queues.get(project_id)
            if queue is None or len(queue) == 0:
                self._is_active_project.discard(project_id)
                continue

            job = queue.popleft()
            self._total_jobs -= 1

            left = self._project_jobs.get(project_id, 0) - 1
            if left <= 0:
                self._project_jobs.pop(project_id, None)
                self._project_queues.pop(project_id, None)
                self._is_active_project.discard(project_id)
            else:
                self._project_jobs[project_id] = left
                self._active_projects.append(project_id)

            return job

        return None

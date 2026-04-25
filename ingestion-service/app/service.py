# pyright: reportMissingImports=false, reportMissingModuleSource=false, reportAttributeAccessIssue=false
import logging
import hashlib
from concurrent import futures

import grpc
from google.protobuf.timestamp_pb2 import Timestamp

from app.config import AppConfig
from app.constants import MAX_INGESTION_BYTES, LogFormat
from app.ingest_worker import IngestJob, IngestLogPayload, IngestWorkerPool, QueueFullError, ProjectQueueFullError, PreparedLogEntry
from app.kafka_batcher import KafkaPublishBatcher
from app.kafka_producer import LogsKafkaProducer
from app.proto import application_pb2, application_pb2_grpc
from app.proto import ingestion_pb2, ingestion_pb2_grpc
from app.proto import logengine_pb2
from app.ml_inference import LogAttributeExtractor, ensure_model_downloaded
from app.utils import compact_json_dumps, to_attribute_string, now_utc, detect_format, parse_json_attributes


class IngestionService(ingestion_pb2_grpc.IngestionServiceServicer):
    def __init__(self, cfg: AppConfig) -> None:
        self._cfg = cfg
        self._app_channel = grpc.insecure_channel(cfg.application_service)
        self._app_client = application_pb2_grpc.ApplicationServiceStub(
            self._app_channel)
        self._logs_producer = LogsKafkaProducer(
            brokers=cfg.kafka_brokers,
            topic=cfg.kafka_topic,
            username=cfg.kafka_username,
            password=cfg.kafka_password,
            mechanism=cfg.kafka_mechanism,
            api_version=cfg.kafka_api_version,
        )
        self._features_producer = LogsKafkaProducer(
            brokers=cfg.kafka_brokers,
            topic=cfg.kafka_features_topic,
            username=cfg.kafka_username,
            password=cfg.kafka_password,
            mechanism=cfg.kafka_mechanism,
            api_version=cfg.kafka_api_version,
        )

        model_path = ensure_model_downloaded(
            model_source=cfg.model_hf_source,
            local_dir=cfg.model_local_dir,
            revision=cfg.model_revision,
            token=cfg.model_hf_token,
        )
        embedding_model_dir = ensure_model_downloaded(
            model_source=cfg.embedding_model_hf_source,
            local_dir=cfg.embedding_model_dir,
            revision="main",
            token=cfg.model_hf_token,
        )
        self._extractor = LogAttributeExtractor(
            model_dir=str(model_path),
            confidence_threshold=cfg.model_confidence_threshold,
            embedding_model_dir=str(embedding_model_dir),
            ner_batch_size=cfg.ml_ner_batch_size,
            embedding_batch_size=cfg.ml_embedding_batch_size,
        )
        logging.info(
            "ML log model enabled from %s (version=%s)",
            model_path,
            self._extractor.model_version,
        )

        self._worker_pool = IngestWorkerPool(
            worker_count=cfg.ingest_worker_count,
            queue_size=cfg.ingest_queue_size,
            per_project_queue_limit=cfg.ingest_per_project_queue_limit,
            overload_high_watermark=cfg.ingest_overload_high_watermark,
            processor=self._process_job,
        )
        self._kafka_batcher = KafkaPublishBatcher(
            max_logs=cfg.ingest_kafka_batch_max_logs,
            max_bytes=cfg.ingest_kafka_batch_max_bytes,
            max_wait_ms=cfg.ingest_kafka_batch_max_wait_ms,
            publish_logs=self._logs_producer.publish,
            publish_features=self._features_producer.publish,
        )

        logging.info(
            "Background ingestion workers started: count=%s queue_size=%s",
            self._worker_pool.worker_count,
            self._worker_pool.queue_size,
        )

    def close(self) -> None:
        self._worker_pool.close()
        self._kafka_batcher.close()
        self._logs_producer.close()
        self._features_producer.close()
        self._app_channel.close()

    def IngestLogs(self, request: ingestion_pb2.IngestLogRequest, context: grpc.ServicerContext) -> ingestion_pb2.IngestLogResponse:
        if len(request.logs) == 0:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT,
                          "Batch must contain at least one log")

        total_size = 0
        payloads: list[IngestLogPayload] = []
        for i, item in enumerate(request.logs):
            if not item.raw_data:
                context.abort(grpc.StatusCode.INVALID_ARGUMENT,
                              f"logs[{i}].raw_data is required")
            total_size += len(item.raw_data)

            if total_size > MAX_INGESTION_BYTES:
                context.abort(
                    grpc.StatusCode.RESOURCE_EXHAUSTED,
                    f"Batch payload exceeds max allowed size: {MAX_INGESTION_BYTES} bytes",
                )
            payloads.append(
                IngestLogPayload(
                    raw_data=item.raw_data,
                )
            )

        try:
            app_info = self._app_client.ValidateApiKey(
                application_pb2.ValidateApiKeyRequest(api_key=request.api_key))
        except grpc.RpcError as exc:
            context.abort(exc.code(), exc.details() or "ValidateApiKey failed")

        try:
            self._worker_pool.enqueue(
                IngestJob(
                    application_id=app_info.application_id,
                    application_name=app_info.application_name,
                    project_id=app_info.project_id,
                    source_ip=request.source_ip,
                    logs=payloads,
                )
            )
        except ProjectQueueFullError:
            context.set_trailing_metadata(
                (("retry-after-ms", str(self._cfg.ingest_retry_after_ms)),)
            )
            context.abort(
                grpc.StatusCode.RESOURCE_EXHAUSTED,
                "Project ingestion quota exceeded, retry later",
            )
        except QueueFullError:
            context.set_trailing_metadata(
                (("retry-after-ms", str(self._cfg.ingest_retry_after_ms)),)
            )
            context.abort(
                grpc.StatusCode.RESOURCE_EXHAUSTED,
                "Ingestion queue is full, retry later",
            )

        return ingestion_pb2.IngestLogResponse(success=True, accepted_count=len(request.logs))

    def _process_job(self, job: IngestJob) -> None:
        logs: list[logengine_pb2.LogItem] = []
        features: list[logengine_pb2.LogFeatureItem] = []

        prepared: list[PreparedLogEntry] = []
        text_indexes: list[int] = []
        text_values: list[str] = []

        for item in job.logs:
            raw_data = item.raw_data
            fmt = detect_format(raw_data)
            attrs: dict[str, str] = {}
            embedding: list[float] = []
            text_value = raw_data.decode("utf-8", errors="replace")

            if fmt == LogFormat.FORMAT_JSON:
                attrs = parse_json_attributes(raw_data)
            elif fmt == LogFormat.FORMAT_TEXT:
                text_indexes.append(len(prepared))
                text_values.append(text_value)

            prepared.append(
                PreparedLogEntry(
                    raw_data=raw_data,
                    fmt=fmt,
                    attrs=attrs,
                    embedding=embedding,
                )
            )

        if text_values:
            predictions = self._extractor.predict_batch(text_values)
            embeddings = self._extractor.encode_embeddings(text_values)

            for idx, prepared_index in enumerate(text_indexes):
                prediction = predictions[idx]
                attrs = {
                    key: to_attribute_string(value)
                    for key, value in prediction.attributes.items()
                }
                attrs["ml_confidence"] = f"{prediction.confidence:.6f}"
                if prediction.low_confidence_attributes:
                    attrs["ml_low_confidence_attributes"] = compact_json_dumps(
                        prediction.low_confidence_attributes,
                    )

                prepared[prepared_index].attrs = attrs
                prepared[prepared_index].embedding = embeddings[idx]

        for entry in prepared:
            raw_data = entry.raw_data
            fmt = entry.fmt
            attrs = entry.attrs
            embedding = entry.embedding

            now = now_utc()
            received_at = Timestamp()
            received_at.FromDatetime(now)

            logs.append(
                logengine_pb2.LogItem(
                    application_id=job.application_id,
                    application_name=job.application_name,
                    project_id=job.project_id,
                    format=fmt.value,
                    source_ip=job.source_ip,
                    received_at=received_at,
                    raw_data=raw_data,
                    attributes=attrs,
                )
            )
            features.append(
                logengine_pb2.LogFeatureItem(
                    application_id=job.application_id,
                    project_id=job.project_id,
                    source_ip=job.source_ip,
                    received_at=received_at,
                    embedding=embedding,
                    attributes=attrs,
                    raw_data_sha256=hashlib.sha256(raw_data).hexdigest(),
                )
            )

        self._kafka_batcher.enqueue(logs, features)


def serve(cfg: AppConfig) -> None:
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=16))
    service = IngestionService(cfg)
    ingestion_pb2_grpc.add_IngestionServiceServicer_to_server(service, server)

    server.add_insecure_port(f"[::]:{cfg.port}")
    server.start()
    logging.info("ingestion-service started on port %s", cfg.port)

    try:
        server.wait_for_termination()
    finally:
        service.close()

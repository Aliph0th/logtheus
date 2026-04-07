# pyright: reportMissingImports=false, reportMissingModuleSource=false, reportAttributeAccessIssue=false
import logging
from concurrent import futures

import grpc
from google.protobuf.timestamp_pb2 import Timestamp

from app.config import AppConfig
from app.constants import MAX_INGESTION_BYTES
from app.ingest_worker import IngestJob, IngestWorkerPool, QueueFullError
from app.kafka_producer import LogsKafkaProducer
from app.parsing import detect_format, parse_attributes, now_utc
from app.proto import application_pb2, application_pb2_grpc
from app.proto import ingestion_pb2, ingestion_pb2_grpc
from app.proto import logengine_pb2
from app.utils.serialization import compact_json_dumps, to_attribute_string


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

        self._extractor = None
        if cfg.model_hf_source:
            from app.ml_inference import LogAttributeExtractor, ensure_model_downloaded

            model_path = ensure_model_downloaded(
                model_source=cfg.model_hf_source,
                local_dir=cfg.model_local_dir,
                revision=cfg.model_revision,
                token=cfg.model_hf_token,
            )
            self._extractor = LogAttributeExtractor(
                model_dir=str(model_path),
                confidence_threshold=cfg.model_confidence_threshold,
            )
            logging.info(
                "ML log model enabled from %s (version=%s)",
                model_path,
                self._extractor.model_version,
            )
        else:
            logging.info(
                "ML log model disabled: LOG_MODEL_HF_SOURCE is not configured")

        self._worker_pool = IngestWorkerPool(
            worker_count=cfg.ingest_worker_count,
            queue_size=cfg.ingest_queue_size,
            processor=self._process_job,
        )

        logging.info(
            "Background ingestion workers started: count=%s queue_size=%s",
            self._worker_pool.worker_count,
            self._worker_pool.queue_size,
        )

    def close(self) -> None:
        self._worker_pool.close()
        self._logs_producer.close()
        self._app_channel.close()

    def IngestLogs(self, request: ingestion_pb2.IngestLogRequest, context: grpc.ServicerContext) -> ingestion_pb2.IngestLogResponse:
        if len(request.logs) == 0:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT,
                          "Batch must contain at least one log")

        total_size = 0
        for i, item in enumerate(request.logs):
            if not item.raw_data:
                context.abort(grpc.StatusCode.INVALID_ARGUMENT,
                              f"logs[{i}].raw_data is required")
            total_size += len(item.raw_data)

        if total_size > MAX_INGESTION_BYTES:
            context.abort(
                grpc.StatusCode.RESOURCE_EXHAUSTED,
                f"Batch payload exceeds max allowed size: {MAX_INGESTION_BYTES} bytes, received: {total_size} bytes",
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
                    raw_logs=[item.raw_data for item in request.logs],
                )
            )
        except QueueFullError:
            context.abort(
                grpc.StatusCode.RESOURCE_EXHAUSTED,
                "Ingestion queue is full, retry later",
            )

        return ingestion_pb2.IngestLogResponse(success=True, accepted_count=len(request.logs))

    def _process_job(self, job: IngestJob) -> None:
        logs: list[logengine_pb2.LogItem] = []

        for raw_data in job.raw_logs:
            fmt = detect_format(raw_data)
            attrs, _ = parse_attributes(raw_data, fmt)

            if self._extractor is not None:
                text = raw_data.decode("utf-8", errors="replace")
                prediction = self._extractor.predict(text)
                ml_attrs = {
                    key: to_attribute_string(value)
                    for key, value in prediction.attributes.items()
                }
                attrs = {**attrs, **ml_attrs}
                attrs["ml_confidence"] = f"{prediction.confidence:.6f}"
                if prediction.low_confidence_attributes:
                    attrs["ml_low_confidence_attributes"] = compact_json_dumps(
                        prediction.low_confidence_attributes,
                    )

            ts = Timestamp()
            ts.FromDatetime(now_utc())

            logs.append(
                logengine_pb2.LogItem(
                    application_id=job.application_id,
                    application_name=job.application_name,
                    project_id=job.project_id,
                    format=fmt.value,
                    source_ip=job.source_ip,
                    received_at=ts,
                    raw_data=raw_data,
                    attributes=attrs,
                )
            )

        payload = logengine_pb2.SaveLogsRequest(logs=logs).SerializeToString()
        self._logs_producer.publish(payload)


def serve(cfg: AppConfig) -> None:
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=16))
    svc = IngestionService(cfg)
    ingestion_pb2_grpc.add_IngestionServiceServicer_to_server(svc, server)

    server.add_insecure_port(f"[::]:{cfg.port}")
    server.start()
    logging.info("ingestion-service started on port %s", cfg.port)

    try:
        server.wait_for_termination()
    finally:
        svc.close()

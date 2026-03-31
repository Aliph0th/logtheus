# pyright: reportMissingImports=false, reportMissingModuleSource=false, reportAttributeAccessIssue=false
import logging
from concurrent import futures

import grpc
from google.protobuf.timestamp_pb2 import Timestamp

from app.config import AppConfig
from app.constants import MAX_INGESTION_BYTES
from app.kafka_producer import LogsKafkaProducer
from app.parsing import detect_format, parse_attributes, now_utc
from app.proto import application_pb2, application_pb2_grpc
from app.proto import ingestion_pb2, ingestion_pb2_grpc
from app.proto import logengine_pb2


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

    def close(self) -> None:
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

        logs = []
        for item in request.logs:
            fmt = detect_format(item.raw_data)
            attrs, _ = parse_attributes(item.raw_data, fmt)

            ts = Timestamp()
            ts.FromDatetime(now_utc())

            logs.append(
                logengine_pb2.LogItem(
                    application_id=app_info.application_id,
                    application_name=app_info.application_name,
                    project_id=app_info.project_id,
                    format=fmt.value,
                    source_ip=request.source_ip,
                    received_at=ts,
                    raw_data=item.raw_data,
                    attributes=attrs,
                )
            )

        payload = logengine_pb2.SaveLogsRequest(logs=logs).SerializeToString()
        try:
            self._logs_producer.publish(payload)
        except Exception as exc:
            logging.exception("Failed to publish logs batch to Kafka")
            context.abort(grpc.StatusCode.INTERNAL,
                          f"Kafka publish failed: {exc}")

        return ingestion_pb2.IngestLogResponse(success=True, accepted_count=len(logs))


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

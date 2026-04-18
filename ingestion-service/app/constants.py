from enum import Enum


MAX_INGESTION_BYTES = 10 * 1024 * 1024


class LogFormat(str, Enum):
    FORMAT_UNKNOWN = "FORMAT_UNKNOWN"
    FORMAT_JSON = "FORMAT_JSON"
    FORMAT_TEXT = "FORMAT_TEXT"
    FORMAT_BINARY = "FORMAT_BINARY"


CANONICAL_ATTRIBUTE_KEYS = (
    "service",
    "level",
    "timestamp",
    "environment",
    "event",
    "error_message",
    "status_code",
    "duration",
    "ip",
    "method",
    "path",
    "useragent",
    "hostname",
)

ATTRIBUTE_ALIASES: dict[str, tuple[str, ...]] = {
    "service": ("service", "svc", "service_name", "app", "application"),
    "level": ("level", "lvl", "severity", "loglevel", "log_level"),
    "timestamp": (
        "timestamp",
        "time",
        "ts",
        "datetime",
        "event_time",
        "event_timestamp",
    ),
    "environment": ("environment", "env", "stage"),
    "event": ("event", "message", "msg", "action"),
    "error_message": (
        "error_message",
        "error",
        "err",
        "errmsg",
        "exception",
        "stacktrace",
    ),
    "status_code": (
        "status_code",
        "status",
        "statuscode",
        "http_status",
        "response_code",
        "code",
    ),
    "duration": (
        "duration",
        "duration_ms",
        "latency",
        "latency_ms",
        "elapsed",
        "response_time",
    ),
    "ip": ("ip", "ip_address", "source_ip", "client_ip", "remote_ip", "remote_addr"),
    "method": ("method", "http_method", "verb"),
    "path": ("path", "url", "uri", "endpoint", "route", "request_path"),
    "useragent": (
        "useragent",
        "user_agent",
        "user-agent",
        "ua",
        "http_user_agent",
    ),
    "hostname": ("hostname", "host", "server", "machine", "node"),
}

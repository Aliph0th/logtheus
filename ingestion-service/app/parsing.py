import json
from datetime import datetime, timezone
from typing import Any, Dict, Tuple

from app.constants import LogFormat
from app.utils.serialization import to_attribute_string


def detect_format(raw: bytes) -> LogFormat:
    try:
        json.loads(raw.decode("utf-8"))
        return LogFormat.FORMAT_JSON
    except Exception:
        pass

    try:
        raw.decode("utf-8")
        return LogFormat.FORMAT_TEXT
    except UnicodeDecodeError:
        return LogFormat.FORMAT_BINARY


def parse_attributes(raw: bytes, fmt: LogFormat) -> Tuple[Dict[str, str], str]:
    if fmt == LogFormat.FORMAT_JSON:
        payload = json.loads(raw.decode("utf-8"))
        if isinstance(payload, dict):
            attrs = {str(k): to_attribute_string(v, ensure_ascii=False)
                     for k, v in payload.items()}
            message = to_attribute_string(
                payload.get("message") or payload.get(
                    "msg") or payload.get("log") or "",
                ensure_ascii=False,
            )
            return attrs, message
        payload_str = to_attribute_string(payload, ensure_ascii=False)
        return {"value": payload_str}, payload_str

    if fmt == LogFormat.FORMAT_TEXT:
        return {}, ""

    return {"raw_size": str(len(raw))}, ""


def extract_api_key_signature(api_key: str) -> str:
    parts = api_key.split("_")
    if len(parts) >= 3:
        return parts[1]
    return ""


def now_utc() -> datetime:
    return datetime.now(timezone.utc)

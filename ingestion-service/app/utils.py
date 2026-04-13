import json
from typing import Any, Dict
from datetime import datetime, timezone
from app.constants import LogFormat


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


def parse_json_attributes(raw: bytes) -> Dict[str, str]:
    payload = json.loads(raw.decode("utf-8"))
    attrs = {str(k): to_attribute_string(v, ensure_ascii=False)
                for k, v in payload.items()}
    return attrs


def compact_json_dumps(value: Any, ensure_ascii: bool = True) -> str:
    return json.dumps(value, ensure_ascii=ensure_ascii, separators=(",", ":"))


def to_attribute_string(value: Any, ensure_ascii: bool = True) -> str:
    if isinstance(value, str):
        return value

    if isinstance(value, (dict, list, tuple, bool)) or value is None:
        return compact_json_dumps(value, ensure_ascii=ensure_ascii)

    return str(value)


def now_utc() -> datetime:
    return datetime.now(timezone.utc)

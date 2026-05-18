import json
from typing import Any, Dict
from datetime import datetime, timezone
from app.constants import LogFormat
from app.constants import ATTRIBUTE_ALIASES


def _normalize_attribute_key(raw_key: str) -> str:
    return "".join(ch for ch in raw_key.lower() if ch.isalnum())


_NORMALIZED_KEY_TO_CANONICAL: dict[str, str] = {
    _normalize_attribute_key(alias): canonical
    for canonical, aliases in ATTRIBUTE_ALIASES.items()
    for alias in aliases
}


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

    if not isinstance(payload, dict):
        return {"event": to_attribute_string(payload, ensure_ascii=False)}

    attrs: Dict[str, str] = {}
    for raw_key, raw_value in payload.items():
        key = str(raw_key)
        canonical_key = _to_canonical_attribute_key(key)
        target_key = canonical_key or key
        value = to_attribute_string(raw_value, ensure_ascii=False)

        if canonical_key is not None and target_key in attrs and key.lower() != canonical_key:
            continue

        attrs[target_key] = value

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


def _to_canonical_attribute_key(raw_key: str) -> str | None:
    return _NORMALIZED_KEY_TO_CANONICAL.get(_normalize_attribute_key(raw_key))

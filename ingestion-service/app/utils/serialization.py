import json
from typing import Any


def compact_json_dumps(value: Any, ensure_ascii: bool = True) -> str:
    return json.dumps(value, ensure_ascii=ensure_ascii, separators=(",", ":"))


def to_attribute_string(value: Any, ensure_ascii: bool = True) -> str:
    if isinstance(value, str):
        return value

    if isinstance(value, (dict, list, tuple, bool)) or value is None:
        return compact_json_dumps(value, ensure_ascii=ensure_ascii)

    return str(value)

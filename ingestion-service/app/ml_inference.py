from __future__ import annotations

import json
import logging
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

import torch
from huggingface_hub import snapshot_download
from transformers import AutoModelForTokenClassification, AutoTokenizer
from sentence_transformers import SentenceTransformer


@dataclass
class ModelPrediction:
    attributes: dict[str, Any]
    low_confidence_attributes: dict[str, Any]
    attribute_confidence: dict[str, Any]
    message: str
    confidence: float
    model_version: str


class LogAttributeExtractor:
    def __init__(self, model_dir: str, confidence_threshold: float, embedding_model_dir: str) -> None:
        self.model_dir = Path(model_dir)
        self.confidence_threshold = confidence_threshold

        self.tokenizer = AutoTokenizer.from_pretrained(
            self.model_dir,
            local_files_only=True,
        )
        self.model = AutoModelForTokenClassification.from_pretrained(
            self.model_dir,
            local_files_only=True,
        )
        self.model.eval()
        self.id2label = self.model.config.id2label
        self.model_version = _read_model_version(self.model_dir)

        self.embedding_model = SentenceTransformer(embedding_model_dir, local_files_only=True)

    def predict(self, text: str) -> ModelPrediction:
        encoded = self.tokenizer(
            text,
            return_tensors="pt",
            truncation=True,
            max_length=256,
            return_offsets_mapping=True,
        )
        offsets = encoded.pop("offset_mapping")[0].tolist()

        with torch.no_grad():
            output = self.model(**encoded, output_hidden_states=True)
        logits = output.logits[0]
        probs = torch.softmax(logits, dim=-1)
        pred_ids = torch.argmax(probs, dim=-1).tolist()
        pred_scores = torch.max(probs, dim=-1).values.tolist()

        grouped_values: dict[str, list[str]] = defaultdict(list)
        grouped_scores: dict[str, list[float]] = defaultdict(list)

        current_label: str | None = None
        current_start: int | None = None
        current_end: int | None = None
        current_scores: list[float] = []

        def flush_current() -> None:
            nonlocal current_label, current_start, current_end, current_scores
            if current_label is None or current_start is None or current_end is None:
                current_label = None
                current_start = None
                current_end = None
                current_scores = []
                return

            value = text[current_start:current_end]
            if value:
                grouped_values[current_label].append(value)
                grouped_scores[current_label].append(
                    float(sum(current_scores) / len(current_scores)
                          ) if current_scores else 0.0
                )

            current_label = None
            current_start = None
            current_end = None
            current_scores = []

        for pred_id, score, (start, end) in zip(pred_ids, pred_scores, offsets):
            if start == end:
                continue

            label = self.id2label[int(pred_id)]
            if label == "O":
                flush_current()
                continue

            prefix, entity = label.split("-", maxsplit=1)

            if prefix == "B" or (current_label is not None and current_label != entity):
                flush_current()
                current_label = entity
                current_start = int(start)
                current_end = int(end)
                current_scores = [float(score)]
            else:
                if current_label is None:
                    current_label = entity
                    current_start = int(start)
                    current_end = int(end)
                    current_scores = [float(score)]
                    continue
                current_label = entity
                current_end = int(end)
                current_scores.append(float(score))

        flush_current()

        attributes: dict[str, Any] = {}
        low_confidence_attributes: dict[str, Any] = {}
        attribute_confidence: dict[str, Any] = {}
        all_scores: list[float] = []

        for label, values in grouped_values.items():
            label_scores = grouped_scores[label]
            mean_score = float(sum(label_scores) /
                               len(label_scores)) if label_scores else 0.0
            all_scores.extend(label_scores)

            value: Any = values[0] if len(values) == 1 else values
            confidence_value: Any = label_scores[0] if len(
                label_scores) == 1 else label_scores
            attribute_confidence[label] = confidence_value

            if mean_score >= self.confidence_threshold:
                attributes[label] = value
            else:
                low_confidence_attributes[label] = value

        overall_confidence = float(
            sum(all_scores) / len(all_scores)) if all_scores else 0.0

        return ModelPrediction(
            attributes=attributes,
            low_confidence_attributes=low_confidence_attributes,
            attribute_confidence=attribute_confidence,
            message=text,
            confidence=overall_confidence,
            model_version=self.model_version,
        )

    def encode_embedding(self, text: str) -> list[float]:
        embedding = self.embedding_model.encode(text, convert_to_tensor=True)
        return embedding.to(dtype=torch.float16).tolist()


def ensure_model_downloaded(
    model_source: str,
    local_dir: str,
    revision: str,
    token: str,
) -> Path:
    local_path = Path(local_dir)
    if _is_model_present(local_path):
        return local_path

    repo_id = _extract_repo_id(model_source)
    logging.info(
        "Downloading log model from Hugging Face repo %s into %s",
        repo_id,
        local_path,
    )

    snapshot_download(
        repo_id=repo_id,
        revision=revision,
        local_dir=str(local_path),
        token=token or None,
    )

    if not _is_model_present(local_path):
        raise RuntimeError(
            "Model download completed, but expected model files are missing")

    return local_path


def _read_model_version(model_dir: Path) -> str:
    metadata_file = model_dir / "model_metadata.json"
    if metadata_file.exists():
        data = json.loads(metadata_file.read_text(encoding="utf-8"))
        return str(data.get("model_version", "unknown"))

    return "unknown"


def _is_model_present(path: Path) -> bool:
    if not path.exists() or not path.is_dir():
        return False

    has_config = (path / "config.json").exists()
    has_tokenizer = (
        path / "tokenizer_config.json").exists() or (path / "tokenizer.json").exists()
    has_weights = any(path.glob("*.safetensors")
                      ) or (path / "pytorch_model.bin").exists()
    return has_config and has_tokenizer and has_weights


def _extract_repo_id(model_source: str) -> str:
    src = model_source.strip()
    if not src:
        raise ValueError("LOG_MODEL_HF_SOURCE must not be empty")

    if src.startswith("http://") or src.startswith("https://"):
        parsed = urlparse(src)
        if "huggingface.co" not in parsed.netloc:
            raise ValueError(
                "LOG_MODEL_HF_SOURCE must point to huggingface.co")

        parts = [p for p in parsed.path.split("/") if p]
        if len(parts) < 2:
            raise ValueError(
                "Cannot parse Hugging Face repo id from LOG_MODEL_HF_SOURCE")
        return f"{parts[0]}/{parts[1]}"

    return src

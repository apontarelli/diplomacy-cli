import json
from importlib import resources
from pathlib import Path
from typing import Any


def load(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8") as f:
        raw = json.load(f)
    if not isinstance(raw, dict):
        raise ValueError(f"{path!r} is not a JSON object")
    return raw


def load_variant_json(
    variant: str, submodule: str, filename: str
) -> dict | list:
    pkg = f"diplomacy_cli.data.{variant}.{submodule}"
    resource = resources.files(pkg).joinpath(filename)
    text = resource.read_text(encoding="utf-8")
    data = json.loads(text)
    return data


def save(data: dict | list, path: Path) -> None:
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)

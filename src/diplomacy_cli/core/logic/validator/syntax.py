import re

from ..schema import SyntaxResult


def normalize_order_string(raw: str):
    raw = raw.lower()
    raw = raw.strip()
    raw = re.sub(r"[‒–—―−]", "-", raw)
    raw = re.sub(r"[^a-z0-9\-/\s]+", "", raw)
    raw = re.sub(r"\s+", " ", raw)
    raw = re.sub(r"\s*-\s*", " - ", raw)
    return raw



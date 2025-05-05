from dataclasses import dataclass
from enum import Enum


class OrderType(str, Enum):
    HOLD = "hold"
    MOVE = "move"
    SUPPORT_HOLD = "support_hold"
    SUPPORT_MOVE = "support_move"
    CONVOY = "convoy"
    BUILD = "build"
    DISBAND = "disband"


@dataclass(frozen=True)
class Rules:
    territory_ids: set[str]
    supply_centers: set[str]
    parent_territory: dict[str, str]
    edges: set[tuple[str, str, str]]
    home_centers: dict[str, set[str]]


@dataclass(frozen=True)
class Order:
    origin: str
    order_type: OrderType
    destination: str | None = None
    convoy_origin: str | None = None
    convoy_destination: str | None = None
    support_origin: str | None = None
    support_destination: str | None = None
    unit_type: str | None = None


@dataclass(frozen=True)
class SyntaxResult:
    raw: str
    normalized: str
    valid: bool
    errors: list[str]
    order: Order | None = None


@dataclass(frozen=True)
class SemanticResult:
    raw: str
    normalized: str
    order: Order
    valid: bool
    errors: list[str]


@dataclass(frozen=True)
class ValidationResult:
    raw: str
    normalized: str
    valid: bool
    errors: list[str]
    order: Order | None = None

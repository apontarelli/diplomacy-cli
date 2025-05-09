from dataclasses import dataclass
from enum import Enum
from typing import Any


class OrderType(str, Enum):
    HOLD = "hold"
    MOVE = "move"
    SUPPORT_HOLD = "support_hold"
    SUPPORT_MOVE = "support_move"
    CONVOY = "convoy"
    BUILD = "build"
    DISBAND = "disband"


TerritoryToUnit = dict[str, str]
Counters = dict[str, int]


@dataclass(frozen=True)
class GameState:
    players: dict[str, Any]
    units: dict[str, Any]
    territory_state: dict[str, Any]
    raw_orders: dict[str, Any]
    game_meta: dict[str, Any]


@dataclass(frozen=True)
class LoadedState:
    game: GameState
    territory_to_unit: TerritoryToUnit
    counters: Counters


@dataclass(frozen=True)
class Rules:
    territory_ids: set[str]
    supply_centers: set[str]
    edges: set[tuple[str, str, str]]
    home_centers: dict[str, set[str]]
    parent_to_coast: dict[str, str]
    coast_to_parent: dict[str, str]


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


@dataclass(frozen=True)
class ResolutionResult:
    player_id: str
    order: Order
    outcome: str


@dataclass(frozen=True)
class Mutation:
    type: str
    payload: dict[str, Any]


@dataclass(frozen=True)
class TurnReport:
    invalid_syntax: dict[str, list[SyntaxResult]]
    invalid_semantic: dict[str, list[SemanticResult]]
    resolution: list[ResolutionResult]
    mutations: list[Mutation]
    new_state: LoadedState

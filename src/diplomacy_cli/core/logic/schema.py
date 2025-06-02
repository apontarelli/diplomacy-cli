from dataclasses import dataclass
from enum import Enum, IntEnum
from typing import Any


class Season(IntEnum):
    SPRING = 0
    FALL = 1
    WINTER = 2


class Phase(IntEnum):
    MOVEMENT = 0
    RETREAT = 1
    ADJUSTMENT = 2


class UnitType(str, Enum):
    ARMY = "army"
    FLEET = "fleet"


class OrderType(str, Enum):
    HOLD = "hold"
    MOVE = "move"
    SUPPORT_HOLD = "support_hold"
    SUPPORT_MOVE = "support_move"
    CONVOY = "convoy"
    BUILD = "build"
    DISBAND = "disband"
    RETREAT = "retreat"


class OutcomeType(str, Enum):
    MOVE_SUCCESS = "move_success"
    MOVE_BOUNCED = "move_bounced"
    MOVE_NO_CONVOY = "move_no_convoy"
    SUPPORT_SUCCESS = "support_success"
    SUPPORT_CUT = "support_cut"
    INVALID_SUPPORT = "invalid_support"
    HOLD_SUCCESS = "hold_success"
    CONVOY_SUCCESS = "convoy_success"
    INVALID_CONVOY = "invalid_convoy"
    DISLODGED = "dislodged"
    RETREAT_SUCCESS = "retreat_success"
    RETREAT_FAILED = "retreat_failed"
    BUILD_SUCCESS = "build_success"
    BUILD_ILLEGAL_LOCATION = "build_illegal_location"
    BUILD_NO_CENTER = "build_no_center"
    DISBAND_SUCCESS = "disband_success"
    DISBAND_FAILED = "disband_failed"


class ChangeType(str, Enum):
    MOVE = "move"
    BUILD = "build"
    DISBAND = "disband"
    SET_OWNER = "set_owner"
    ELIMINATE = "eliminate"


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
    dislodged: set[str]


@dataclass(frozen=True)
class Rules:
    territory_ids: set[str]
    supply_centers: set[str]
    home_centers: dict[str, set[str]]
    display_name: dict[str, str]
    territory_type: dict[str, str]
    has_coast: set[str]
    parent_coasts: dict[str, list[str]]
    parent_to_coast: dict[str, dict[str, str]]
    coast_to_parent: dict[str, str]
    edges: set[tuple[str, str, str]]
    adjacency_map: dict[str, list[tuple[str, str]]]


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
    player_id: str
    raw: str
    normalized: str
    valid: bool
    errors: list[str]
    order: Order | None = None


@dataclass(frozen=True)
class SemanticResult:
    player_id: str
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


@dataclass
class ResolutionSoA:
    unit_id: list[str]
    owner_id: list[str]
    unit_type: list[UnitType]
    orig_territory: list[str]
    order_type: list[OrderType]
    move_destination: list[str | None]
    support_origin: list[str | None]
    support_destination: list[str | None]
    convoy_origin: list[str | None]
    convoy_destination: list[str | None]
    new_territory: list[str]
    strength: list[int]
    dislodged: list[bool]
    support_cut: list[bool]
    convoy_path_flat: list[str]
    convoy_path_start: list[int]
    convoy_path_len: list[int]
    outcome: list[OutcomeType | None]


@dataclass(frozen=True)
class ResolutionResult:
    unit_id: str
    owner_id: str
    unit_type: UnitType
    origin_territory: str
    semantic_result: SemanticResult
    outcome: OutcomeType
    resolved_territory: str
    strength: int
    dislodged_by_id: str | None
    destination: str | None
    convoy_path: list[str] | None
    supported_unit_id: str | None
    duplicate_orders: list[SemanticResult]


@dataclass(frozen=True)
class PhaseResolutionReport:
    phase: Phase
    season: Season
    year: int
    syntax_errors: list[SyntaxResult]
    semantic_errors: list[SemanticResult]
    resolution_results: list[ResolutionResult]


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

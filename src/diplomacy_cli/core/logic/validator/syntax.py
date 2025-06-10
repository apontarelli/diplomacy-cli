import re

from ..schema import Order, OrderType, SyntaxResult, UnitType
from ..turn_code import Phase


class ParseError(Exception):
    def __init__(self, message: str):
        super().__init__(message)


def normalize_order_string(raw: str) -> str:
    raw = raw.lower()
    raw = raw.strip()
    raw = re.sub(r"[‒–—―−]", "-", raw)
    raw = re.sub(r"\s*[/]\s*", "_", raw)
    raw = re.sub(r"[^a-z0-9\-_/\s]+", "", raw)
    raw = re.sub(r"\s+", " ", raw)
    raw = re.sub(r"\s*-\s*", " - ", raw)
    return raw


def expect(tokens, tok):
    if not tokens or tokens.pop(0) != tok:
        raise ParseError(f"Expected {tok!r}")


def take_province(tokens: list[str]) -> str:
    if not tokens:
        raise ParseError("Expected province, but got end of input")
    prov = tokens.pop(0)
    return prov


def take_unit_type(tokens: list[str]) -> UnitType:
    if not tokens:
        raise ParseError(
            "Expected unit type (‘army’ or ‘fleet’), but got end of input"
        )
    tok = tokens.pop(0)
    if tok == "army":
        return UnitType.ARMY
    if tok == "fleet":
        return UnitType.FLEET
    raise ParseError(f"Expected unit type ‘army’ or ‘fleet’, but got {tok!r}")


def ensure_no_tokens(tokens: list[str]) -> None:
    if tokens:
        raise ParseError(f"Extra tokens: {tokens}")


def parse_support_move(tokens: list[str]):
    origin = take_province(tokens)
    expect(tokens, "s")
    support_origin = take_province(tokens)
    expect(tokens, "-")
    destination = take_province(tokens)
    return Order(
        origin=origin,
        order_type=OrderType.SUPPORT_MOVE,
        support_destination=destination,
        support_origin=support_origin,
    )


def parse_convoy(tokens: list[str]):
    origin = take_province(tokens)
    expect(tokens, "c")
    convoy_origin = take_province(tokens)
    expect(tokens, "-")
    convoy_destination = take_province(tokens)
    return Order(
        origin=origin,
        order_type=OrderType.CONVOY,
        convoy_origin=convoy_origin,
        convoy_destination=convoy_destination,
    )


def parse_support_hold(tokens: list[str]):
    origin = take_province(tokens)
    expect(tokens, "s")
    support_origin = take_province(tokens)
    return Order(
        origin=origin,
        order_type=OrderType.SUPPORT_HOLD,
        support_origin=support_origin,
    )


def parse_move(tokens: list[str]):
    origin = take_province(tokens)
    expect(tokens, "-")
    destination = take_province(tokens)
    ensure_no_tokens(tokens)
    return Order(
        origin=origin,
        order_type=OrderType.MOVE,
        destination=destination,
    )


def parse_hold(tokens: list[str]):
    origin = take_province(tokens)
    expect(tokens, "hold")
    return Order(origin=origin, order_type=OrderType.HOLD)


def parse_build(tokens: list[str]):
    expect(tokens, "build")
    unit_type = take_unit_type(tokens)
    origin = take_province(tokens)
    return Order(
        origin=origin,
        unit_type=unit_type,
        order_type=OrderType.BUILD,
    )


def parse_disband(tokens: list[str]):
    expect(tokens, "disband")
    unit_type = take_unit_type(tokens)
    origin = take_province(tokens)
    return Order(
        origin=origin,
        unit_type=unit_type,
        order_type=OrderType.DISBAND,
    )


def parse_retreat(tokens: list[str]) -> Order:
    origin = take_province(tokens)
    expect(tokens, "-")
    destination = take_province(tokens)
    ensure_no_tokens(tokens)
    return Order(
        origin=origin,
        destination=destination,
        order_type=OrderType.RETREAT,
    )


MOVEMENT_PARSERS = [
    parse_support_move,
    parse_convoy,
    parse_support_hold,
    parse_move,
    parse_hold,
]

RETREAT_PARSERS = [parse_retreat]

ADJUSTMENT_PARSERS = [parse_build, parse_disband]


def dispatch_parsers(tokens: list[str], phase: Phase):
    if phase == Phase.MOVEMENT:
        candidates = MOVEMENT_PARSERS
    elif phase == Phase.RETREAT:
        candidates = RETREAT_PARSERS
    elif phase == Phase.ADJUSTMENT:
        candidates = ADJUSTMENT_PARSERS
    else:
        raise ParseError(f"Unknown phase: {phase}")

    for parser in candidates:
        try:
            order = parser(tokens.copy())
            return order
        except ParseError:
            continue
    raise ParseError(
        f"Unrecognized order for {phase.name} phase: {' '.join(tokens)}"
    )


def parse_syntax(player_id: str, raw: str, phase: Phase) -> SyntaxResult:
    normalized = normalize_order_string(raw)
    errors: list[str] = []
    order = None

    try:
        tokens = normalized.split(" ")
        order = dispatch_parsers(tokens, phase)
        valid = True
    except ParseError as pe:
        errors.append(str(pe))
        valid = False
    except Exception as e:
        errors.append(f"Internal parse error: {e}")
        valid = False

    return SyntaxResult(
        player_id=player_id,
        raw=raw,
        normalized=normalized,
        valid=valid,
        errors=errors,
        order=order,
    )

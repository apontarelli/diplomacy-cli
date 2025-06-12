from diplomacy_cli.core.logic.schema import (
    Order,
    OrderType,
    OutcomeType,
    Phase,
    PhaseResolutionReport,
    ResolutionResult,
    Season,
    SemanticResult,
    SyntaxResult,
    UnitType,
)


def syntax_result_to_dict(s: SyntaxResult) -> dict:
    return {
        "player_id": s.player_id,
        "raw": s.raw,
        "normalized": s.normalized,
        "valid": s.valid,
        "errors": s.errors,
        "order": order_to_dict(s.order) if s.order else None,
    }


def semantic_result_to_dict(s: SemanticResult) -> dict:
    return {
        "player_id": s.player_id,
        "raw": s.raw,
        "normalized": s.normalized,
        "order": order_to_dict(s.order),
        "valid": s.valid,
        "errors": s.errors,
    }


def resolution_result_to_dict(r: ResolutionResult) -> dict:
    return {
        "unit_id": r.unit_id,
        "owner_id": r.owner_id,
        "unit_type": r.unit_type.value,
        "origin_territory": r.origin_territory,
        "semantic_result": semantic_result_to_dict(r.semantic_result),
        "outcome": r.outcome.name,
        "resolved_territory": r.resolved_territory,
        "strength": r.strength,
        "dislodged_by_id": r.dislodged_by_id,
        "destination": r.destination,
        "convoy_path": r.convoy_path,
        "supported_unit_id": r.supported_unit_id,
        "duplicate_orders": [
            semantic_result_to_dict(s) for s in r.duplicate_orders
        ],
    }


def order_to_dict(order: Order) -> dict:
    return {
        "origin": order.origin,
        "order_type": order.order_type.name,
        "destination": order.destination,
        "convoy_origin": order.convoy_origin,
        "convoy_destination": order.convoy_destination,
        "support_origin": order.support_origin,
        "support_destination": order.support_destination,
        "unit_type": order.unit_type,
    }


def phase_resolution_report_to_dict(report: PhaseResolutionReport) -> dict:
    return {
        "phase": report.phase.name,
        "season": report.season.name,
        "year": report.year,
        "valid_syntax": [syntax_result_to_dict(s) for s in report.valid_syntax],
        "valid_semantics": [
            semantic_result_to_dict(s) for s in report.valid_semantics
        ],
        "syntax_errors": [
            syntax_result_to_dict(s) for s in report.syntax_errors
        ],
        "semantic_errors": [
            semantic_result_to_dict(s) for s in report.semantic_errors
        ],
        "resolution_results": [
            resolution_result_to_dict(r) for r in report.resolution_results
        ],
    }


def order_from_dict(d: dict) -> Order:
    return Order(
        origin=d["origin"],
        order_type=OrderType[d["order_type"]],
        destination=d.get("destination"),
        convoy_origin=d.get("convoy_origin"),
        convoy_destination=d.get("convoy_destination"),
        support_origin=d.get("support_origin"),
        support_destination=d.get("support_destination"),
        unit_type=d.get("unit_type"),
    )


def syntax_result_from_dict(d: dict) -> SyntaxResult:
    return SyntaxResult(
        player_id=d["player_id"],
        raw=d["raw"],
        normalized=d["normalized"],
        valid=d["valid"],
        errors=d["errors"],
        order=order_from_dict(d["order"]) if d["order"] is not None else None,
    )


def semantic_result_from_dict(d: dict) -> SemanticResult:
    return SemanticResult(
        player_id=d["player_id"],
        raw=d["raw"],
        normalized=d["normalized"],
        order=order_from_dict(d["order"]),
        valid=d["valid"],
        errors=d["errors"],
    )


def resolution_result_from_dict(d: dict) -> ResolutionResult:
    return ResolutionResult(
        unit_id=d["unit_id"],
        owner_id=d["owner_id"],
        unit_type=UnitType(d["unit_type"]),
        origin_territory=d["origin_territory"],
        semantic_result=semantic_result_from_dict(d["semantic_result"]),
        outcome=OutcomeType[d["outcome"]],
        resolved_territory=d["resolved_territory"],
        strength=d["strength"],
        dislodged_by_id=d.get("dislodged_by_id"),
        destination=d.get("destination"),
        convoy_path=d.get("convoy_path"),
        supported_unit_id=d.get("supported_unit_id"),
        duplicate_orders=[
            semantic_result_from_dict(s) for s in d["duplicate_orders"]
        ],
    )


def phase_resolution_report_from_dict(d: dict) -> PhaseResolutionReport:
    return PhaseResolutionReport(
        phase=Phase[d["phase"]],
        season=Season[d["season"]],
        year=d["year"],
        valid_syntax=[syntax_result_from_dict(s) for s in d["valid_syntax"]],
        valid_semantics=[
            semantic_result_from_dict(s) for s in d["valid_semantics"]
        ],
        syntax_errors=[syntax_result_from_dict(s) for s in d["syntax_errors"]],
        semantic_errors=[
            semantic_result_from_dict(s) for s in d["semantic_errors"]
        ],
        resolution_results=[
            resolution_result_from_dict(r) for r in d["resolution_results"]
        ],
    )

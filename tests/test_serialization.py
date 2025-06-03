from diplomacy_cli.core.logic.schema import (
    Order,
    OrderType,
    SyntaxResult,
    SemanticResult,
    ResolutionResult,
    UnitType,
    OutcomeType,
    PhaseResolutionReport,
    Phase,
    Season,
)
from diplomacy_cli.core.logic.serialization import (
    order_from_dict,
    order_to_dict,
    syntax_result_from_dict,
    syntax_result_to_dict,
    semantic_result_from_dict,
    semantic_result_to_dict,
    resolution_result_from_dict,
    resolution_result_to_dict,
    phase_resolution_report_from_dict,
    phase_resolution_report_to_dict,
)


def test_order_roundtrip():
    order = Order(
        origin="par",
        order_type=OrderType.MOVE,
        destination="bur",
        convoy_origin=None,
        convoy_destination=None,
        support_origin=None,
        support_destination=None,
        unit_type="A",
    )
    assert order_from_dict(order_to_dict(order)) == order


def test_syntax_result_roundtrip():
    order = Order("lon", OrderType.HOLD)
    syntax = SyntaxResult(
        player_id="ENG",
        raw="lon hold",
        normalized="lon h",
        valid=True,
        errors=[],
        order=order,
    )
    assert syntax_result_from_dict(syntax_result_to_dict(syntax)) == syntax


def test_semantic_result_roundtrip():
    order = Order("bre", OrderType.MOVE, destination="pic")
    semantic = SemanticResult(
        player_id="FRA",
        raw="bre-pic",
        normalized="bre m pic",
        order=order,
        valid=True,
        errors=[],
    )
    assert (
        semantic_result_from_dict(semantic_result_to_dict(semantic)) == semantic
    )


def test_resolution_result_roundtrip():
    sem = SemanticResult(
        player_id="RUS",
        raw="sev-bla",
        normalized="sev m bla",
        order=Order("sev", OrderType.MOVE, destination="bla"),
        valid=True,
        errors=[],
    )
    result = ResolutionResult(
        unit_id="U1",
        owner_id="RUS",
        unit_type=UnitType.FLEET,
        origin_territory="sev",
        semantic_result=sem,
        outcome=OutcomeType.MOVE_SUCCESS,
        resolved_territory="bla",
        strength=1,
        dislodged_by_id=None,
        destination="bla",
        convoy_path=None,
        supported_unit_id=None,
        duplicate_orders=[],
    )
    assert (
        resolution_result_from_dict(resolution_result_to_dict(result)) == result
    )


def test_phase_resolution_report_roundtrip():
    order = Order("rom", OrderType.HOLD, unit_type="A")
    syntax = SyntaxResult("ITA", "rom hold", "rom h", True, [], order)
    semantic = SemanticResult("ITA", "rom hold", "rom h", order, True, [])
    result = ResolutionResult(
        unit_id="U2",
        owner_id="ITA",
        unit_type=UnitType.ARMY,
        origin_territory="rom",
        semantic_result=semantic,
        outcome=OutcomeType.MOVE_SUCCESS,
        resolved_territory="rom",
        strength=1,
        dislodged_by_id=None,
        destination=None,
        convoy_path=None,
        supported_unit_id=None,
        duplicate_orders=[],
    )

    report = PhaseResolutionReport(
        phase=Phase.MOVEMENT,
        season=Season.SPRING,
        year=1901,
        valid_syntax=[syntax],
        valid_semantics=[semantic],
        syntax_errors=[],
        semantic_errors=[],
        resolution_results=[result],
    )

    assert (
        phase_resolution_report_from_dict(
            phase_resolution_report_to_dict(report)
        )
        == report
    )

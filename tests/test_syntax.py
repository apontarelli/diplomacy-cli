import pytest

from diplomacy_cli.core.logic.validator.syntax import (
    Order,
    OrderType,
    ParseError,
    Phase,
    dispatch_parsers,
    ensure_no_tokens,
    expect,
    normalize_order_string,
    parse_build,
    parse_convoy,
    parse_disband,
    parse_hold,
    parse_move,
    parse_support_hold,
    parse_support_move,
    parse_syntax,
    take_province,
    take_unit_type,
)


@pytest.mark.parametrize(
    "raw, expected",
    [
        ("  BUR  S   PAR - PIC! ", "bur s par - pic"),
        ("Par-Pic", "par - pic"),
        (" MUN    hold", "mun hold"),
        ("stp/sc hold", "stp_sc hold"),
        ("bal c pru – den", "bal c pru - den"),
        ("bal c pru — den", "bal c pru - den"),
        ("bal c pru ‒ den", "bal c pru - den"),
        ("par.pic", "parpic"),
        ("bur,s par", "burs par"),
        ("  Par...   -   Pic??", "par - pic"),
        ("stp / sc hold", "stp_sc hold"),
    ],
)
def test_normalize_various_cases(raw, expected):
    assert normalize_order_string(raw) == expected


def test_expect():
    raw = ["build", "army", "stp"]
    expected = None
    assert expect(raw, "build") == expected
    raw = ["army", "stp"]
    with pytest.raises(ParseError) as excinfo:
        expect(raw, "build")
    assert f"Expected {'build'!r}" in str(excinfo.value)


def test_expect_success_and_failure():
    raw = ["build", "army", "stp"]
    assert expect(raw, "build") is None
    assert raw == ["army", "stp"]

    raw = ["army", "stp"]
    with pytest.raises(ParseError) as excinfo:
        expect(raw, "build")
    assert "Expected 'build'" in str(excinfo.value)


def test_take_province_success():
    raw = ["par"]
    prov = take_province(raw)
    assert prov == "par"
    assert raw == []


def test_take_province_empty():
    with pytest.raises(ParseError) as excinfo:
        take_province([])
    assert "Expected province, but got end of input" in str(excinfo.value)


def test_take_unit_type_success():
    raw = ["fleet", "lon"]
    utype = take_unit_type(raw)
    assert utype == "fleet"
    assert raw == ["lon"]


def test_take_unit_type_invalid_and_empty():
    with pytest.raises(ParseError) as excinfo:
        take_unit_type(["boat"])
    assert "Expected unit type ‘army’ or ‘fleet’, but got 'boat'" in str(
        excinfo.value
    )

    with pytest.raises(ParseError) as excinfo2:
        take_unit_type([])
    assert (
        "Expected unit type (‘army’ or ‘fleet’), but got end of input"
        in str(excinfo2.value)
    )


def test_ensure_no_tokens_success_and_failure():
    ensure_no_tokens([])
    with pytest.raises(ParseError) as excinfo:
        ensure_no_tokens(["x", "y"])
    assert "Extra tokens: ['x', 'y']" in str(excinfo.value)


def test_parse_support_move_success_and_missing_token():
    raw = ["par", "s", "mar", "-", "bre"]
    order = parse_support_move(raw.copy())
    assert isinstance(order, Order)
    assert order.origin == "par"
    assert order.order_type == OrderType.SUPPORT_MOVE
    assert order.support_origin == "mar"
    assert order.support_destination == "bre"

    with pytest.raises(ParseError):
        parse_support_move(["par", "x", "mar", "-", "bre"])


def test_parse_support_hold_success_and_missing():
    raw = ["tun", "s", "ven"]
    order = parse_support_hold(raw.copy())
    assert order.origin == "tun"
    assert order.order_type == OrderType.SUPPORT_HOLD
    assert order.support_origin == "ven"

    with pytest.raises(ParseError):
        parse_support_hold(["tun"])


def test_parse_convoy_success_and_missing():
    raw = ["ida", "c", "bul", "-", "con"]
    order = parse_convoy(raw.copy())
    assert order.origin == "ida"
    assert order.order_type == OrderType.CONVOY
    assert order.convoy_origin == "bul"
    assert order.convoy_destination == "con"

    with pytest.raises(ParseError):
        parse_convoy(["ida", "c", "bul", "x", "con"])


def test_parse_move_success_and_extra_and_missing():
    raw = ["loe", "-", "fin"]
    order = parse_move(raw.copy())
    assert order.origin == "loe"
    assert order.order_type == OrderType.MOVE
    assert order.destination == "fin"

    with pytest.raises(ParseError):
        parse_move(["loe", "-", "fin", "x"])

    with pytest.raises(ParseError):
        parse_move(["loe", "x", "fin"])


def test_parse_hold_success_and_missing():
    raw = ["ank", "hold"]
    order = parse_hold(raw.copy())
    assert order.origin == "ank"
    assert order.order_type == OrderType.HOLD

    with pytest.raises(ParseError):
        parse_hold(["ank"])


def test_parse_build_success_and_errors():
    raw = ["build", "army", "ber"]
    order = parse_build(raw.copy())
    assert order.origin == "ber"
    assert order.unit_type == "army"
    assert order.order_type == OrderType.BUILD

    with pytest.raises(ParseError):
        parse_build(["bld", "army", "ber"])
    with pytest.raises(ParseError):
        parse_build(["build"])


def test_parse_disband_success_and_errors():
    raw = ["disband", "fleet", "sev"]
    order = parse_disband(raw.copy())
    assert order.origin == "sev"
    assert order.unit_type == "fleet"
    assert order.order_type == OrderType.DISBAND

    with pytest.raises(ParseError):
        parse_disband(["dband", "fleet", "sev"])
    with pytest.raises(ParseError):
        parse_disband(["disband"])


def test_dispatch_parsers_success():
    tokens = ["mar", "-", "ber"]
    order = dispatch_parsers(tokens, Phase.MOVEMENT)
    assert order.order_type == OrderType.MOVE
    assert order.origin == "mar"
    assert order.destination == "ber"


def test_dispatch_parsers_failure():
    with pytest.raises(ParseError) as excinfo:
        dispatch_parsers(["foo", "bar"], Phase.MOVEMENT)
    assert "Unrecognized order for MOVEMENT phase: foo bar" in str(
        excinfo.value
    )


def test_parse_syntax_valid_move():
    raw = "mar - ber"
    player_id = "eng"
    result = parse_syntax(player_id, raw, Phase.MOVEMENT)
    assert result.raw == raw
    assert result.valid is True
    assert result.errors == []
    assert result.order is not None
    assert result.order.order_type == OrderType.MOVE


def test_parse_syntax_invalid():
    raw = "invalid order"
    player_id = "eng"
    result = parse_syntax(player_id, raw, Phase.MOVEMENT)
    assert result.raw == raw
    assert result.valid is False
    assert result.order is None
    assert len(result.errors) == 1
    assert (
        "Unrecognized order for MOVEMENT phase: invalid order"
        in result.errors[0]
    )

from diplomacy_cli.core.logic.state import load_orders

from collections import defaultdict


def test_load_orders_returns_defaultdict_list():
    result = load_orders()

    assert isinstance(result, defaultdict)
    assert result.default_factory is list
    assert dict(result) == {}


def test_load_orders_supports_appending_orders():
    orders = load_orders()
    orders["ENG"].append("ENG - NTH")
    orders["ENG"].append("ENG - NWG")

    assert orders["ENG"] == ["ENG - NTH", "ENG - NWG"]

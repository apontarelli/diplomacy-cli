from typing import cast

from .schema import Rules
from .storage import load_variant_json


def load_rules(variant):
    territories_raw = load_variant_json(variant, "world", "territories.json")
    assert isinstance(territories_raw, dict)
    territories = cast(dict[str, dict], territories_raw)

    edge_data = load_variant_json(variant, "world", "edges.json")

    territory_ids = set()
    supply_centers = set()
    parent_territory = {}
    edges = set()
    home_centers = {}

    for tid, data in territories.items():
        territory_ids.add(tid)
        if data.get("is_supply_center"):
            supply_centers.add(tid)
        if data.get("parent"):
            parent_territory[tid] = data["parent"]
        if data.get("home_center"):
            home_centers.setdefault(data["home_center"], set()).add(data)

    for edge in edge_data:
        a, b, mode = edge["from"], edge["to"], edge["mode"]
        edges.add((a, b, mode))
        edges.add((b, a, mode))

    return Rules(
        territory_ids=territory_ids,
        supply_centers=supply_centers,
        parent_territory=parent_territory,
        edges=edges,
        home_centers=home_centers,
    )

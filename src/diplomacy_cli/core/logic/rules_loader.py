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
    edges = set()
    home_centers = {}
    parent_to_coast = {}
    coast_to_parent = {}

    for tid, data in territories.items():
        territory_ids.add(tid)
        if data.get("is_supply_center"):
            supply_centers.add(tid)
        coasts = data.get("coasts")
        if coasts:
            for coast in coasts:
                coast_id = f"{tid}_{coast}"
                territory_ids.add(coast_id)
                parent_to_coast.setdefault(tid, {})[coast] = coast_id
                coast_to_parent[coast_id] = tid

        home_country = data.get("home_country")
        if home_country:
            home_centers.setdefault(home_country, set()).add(tid)

    for edge in edge_data:
        a, b, mode = edge["from"], edge["to"], edge["mode"]
        edges.add((a, b, mode))
        edges.add((b, a, mode))

    return Rules(
        territory_ids=territory_ids,
        supply_centers=supply_centers,
        edges=edges,
        home_centers=home_centers,
        parent_to_coast=parent_to_coast,
        coast_to_parent=coast_to_parent,
    )

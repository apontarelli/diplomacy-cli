from collections import defaultdict

from .schema import Rules
from .storage import load_variant_json


def load_rules(variant) -> Rules:
    territories_raw = load_variant_json(variant, "world", "territories.json")
    edge_data = load_variant_json(variant, "world", "edges.json")
    nation_data = load_variant_json(variant, "world", "nations.json")
    assert isinstance(territories_raw, dict)
    territory_ids = set()
    supply_centers = set()
    home_centers = defaultdict(set)

    territory_display_names = {}
    nation_display_names = {}
    territory_type = {}
    has_coast = set()
    parent_coasts = defaultdict(list)
    parent_to_coast = {}
    coast_to_parent = {}

    for tid, data in territories_raw.items():
        territory_ids.add(tid)
        territory_display_names[tid] = data["display_name"]
        territory_type[tid] = data["type"]  # "land" / "sea" / "coast"
        if data.get("is_supply_center"):
            supply_centers.add(tid)
        if data.get("has_coast"):
            has_coast.add(tid)
        if data.get("home_country"):
            home_centers[data["home_country"]].add(tid)

        for coast in data.get("coasts", []):
            coast_id = f"{tid}_{coast}"
            territory_ids.add(coast_id)
            territory_type[coast_id] = "coast"
            territory_display_names[coast_id] = (
                f"{data['display_name']} ({coast.upper()})"
            )
            parent_coasts[tid].append(coast_id)

            parent_to_coast.setdefault(tid, []).append(coast_id)
            coast_to_parent[coast_id] = tid

    for nation in nation_data:
        nation_display_names[nation["id"]] = nation["display_name"]

    edges = set()
    adjacency_map = defaultdict(list)
    for e in edge_data:
        a, b, mode = e["from"], e["to"], e["mode"]
        for src, dst in ((a, b), (b, a)):
            edges.add((src, dst, mode))
            adjacency_map[src].append((dst, mode))

    return Rules(
        territory_ids=territory_ids,
        supply_centers=supply_centers,
        home_centers=dict(home_centers),
        territory_display_names=territory_display_names,
        nation_display_names=nation_display_names,
        territory_type=territory_type,
        has_coast=has_coast,
        parent_coasts=dict(parent_coasts),
        edges=edges,
        adjacency_map=dict(adjacency_map),
        parent_to_coast=parent_to_coast,
        coast_to_parent=coast_to_parent,
    )

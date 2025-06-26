from diplomacy_cli.core.logic.rules_loader import Rules, load_rules


def test_load_rules_smoke():
    rules = load_rules("classic")
    assert isinstance(rules, Rules)


def test_load_rules_structure():
    rules = load_rules("classic")

    # Core sets and mappings
    assert isinstance(rules.territory_ids, set)
    assert len(rules.territory_ids) > 0
    assert isinstance(rules.supply_centers, set)
    assert len(rules.supply_centers) > 0
    assert isinstance(rules.home_centers, dict)
    assert len(rules.home_centers) > 0

    # Coast mappings
    assert isinstance(rules.parent_to_coast, dict)
    assert len(rules.parent_to_coast) > 0
    assert isinstance(rules.coast_to_parent, dict)
    assert len(rules.coast_to_parent) > 0
    # parent_coasts should map each parent territory to a list of its coast IDs
    assert isinstance(rules.parent_coasts, dict)
    for parent, coasts in rules.parent_coasts.items():
        assert parent in rules.territory_ids
        assert isinstance(coasts, list)
        assert all(isinstance(c, str) for c in coasts)

    # Territory metadata columns
    assert isinstance(rules.territory_display_names, dict)
    # Every territory ID should have a display name
    assert set(rules.territory_display_names.keys()) == rules.territory_ids

    assert isinstance(rules.territory_type, dict)
    # Types should be one of the expected values
    valid_types = {"land", "sea", "coast"}
    assert set(rules.territory_type.keys()) == rules.territory_ids
    assert set(rules.territory_type.values()).issubset(valid_types)

    assert isinstance(rules.has_coast, set)
    # has_coast must be a subset of territory_ids
    assert rules.has_coast.issubset(rules.territory_ids)

    # Adjacency structures
    assert isinstance(rules.edges, set)
    assert len(rules.edges) > 0

    assert isinstance(rules.adjacency_map, dict)
    for src, neighbors in rules.adjacency_map.items():
        assert isinstance(src, str)
        assert src in rules.territory_ids
        assert isinstance(neighbors, list)
        for dst, mode in neighbors:
            assert isinstance(dst, str)
            assert dst in rules.territory_ids
            assert mode in ("land", "sea", "both")

    # Coast lookup round-trips
    for parent, coast_ids in rules.parent_to_coast.items():
        assert isinstance(parent, str)
        assert isinstance(coast_ids, list)
        assert coast_ids, f"{parent} should have at least one coast"
        for coast_id in coast_ids:
            assert isinstance(coast_id, str)
            assert rules.coast_to_parent[coast_id] == parent


def test_edges_are_bidirectional():
    rules = load_rules("classic")
    for a, b, m in rules.edges:
        assert (b, a, m) in rules.edges


def test_home_centers_contents():
    rules = load_rules("classic")
    for country, centers in rules.home_centers.items():
        assert isinstance(country, str)
        assert isinstance(centers, set)
        assert all(isinstance(t, str) for t in centers)

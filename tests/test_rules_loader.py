from diplomacy_cli.core.logic.rules_loader import Rules, load_rules


def test_load_rules_smoke():
    rules = load_rules("classic")
    assert isinstance(rules, Rules)


def test_load_rules_structure():
    rules = load_rules("classic")
    assert isinstance(rules.territory_ids, set)
    assert len(rules.territory_ids) > 0
    assert isinstance(rules.supply_centers, set)
    assert len(rules.supply_centers) > 0
    assert isinstance(rules.parent_to_coast, dict)
    assert len(rules.parent_to_coast) > 0
    assert isinstance(rules.coast_to_parent, dict)
    assert len(rules.coast_to_parent) > 0
    assert isinstance(rules.edges, set)
    assert len(rules.edges) > 0
    assert isinstance(rules.home_centers, dict)


def test_edges_are_bidirectional():
    rules = load_rules("classic")
    for a, b, m in rules.edges:
        assert (b, a, m) in rules.edges


def test_home_centers_contents():
    rules = load_rules("classic")
    for key, value in rules.home_centers.items():
        assert isinstance(key, str)
        assert isinstance(value, set)
        assert all(isinstance(v, str) for v in value)

from collections import defaultdict


def format_state(state):
    game = format_game(state["game"])
    players = format_players(state["players"])
    territory_state = format_territory_state(state["territory_state"])
    units = format_units(state["units"])

    output = []
    output.append(f"\n{game}")
    output.append(f"\n{players}")
    output.append(f"\n{territory_state}")
    output.append(f"\n{units}")

    return "\n".join(output)


def format_game(game):
    output = []
    output.append(f"--- Game: {game['game_id']} ---")
    output.append(f"Variant: {game['variant']}")
    output.append(f"Turn: {game['turn_code']}")
    output.append(f"Status: {game['status']}")

    return "\n".join(output)


def format_players(players):
    output = []
    output.append("--- Players ---")
    for k, v in players.items():
        output.append(f"{k} - {v['status']}")

    return "\n".join(output)


def format_territory_state(territory_state):
    output = []
    by_owner = group_territory_state_by_owner(territory_state)

    output.append("--- Territories ---")
    for owner, owned_territories in by_owner.items():
        output.append(f"\n{owner}")
        for territory_id in owned_territories:
            output.append(f" - {territory_id}")

    return "\n".join(output)


def format_units(units):
    output = []
    by_owner = group_units_by_owner(flatten_units(units))

    output.append("--- Units ---")
    for owner, units in by_owner.items():
        output.append(f"\n{owner}")
        for u in units:
            output.append(f" - {u['type']} - {u['territory_id']}")

    return "\n".join(output)


def flatten_units(units_dict):
    return [
        {"territory_id": territory, **unit}
        for territory, unit in units_dict.items()
    ]


def group_territory_state_by_owner(territory_state_dict):
    grouped = defaultdict(list)
    for territory_id, data in territory_state_dict.items():
        owner = data.get("owner_id", "None")
        grouped[owner].append(territory_id)
    return grouped


def group_units_by_owner(unit_list):
    grouped = defaultdict(list)
    for unit in unit_list:
        owner = unit.get("owner_id", "None")
        grouped[owner].append(unit)
    return grouped

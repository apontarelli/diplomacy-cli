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
    output.append(f"--- Game: {game["game_id"]} ---")
    output.append(f"Variant: {game["variant"]}")
    output.append(f"Turn: {game["turn"]} | {game["season"]} | {game["phase"]}")
    output.append(f"Status: {game["status"]}")

    return "\n".join(output)

def format_players(players):
    output = []
    output.append(f"--- Players ---")
    for p in players:
        output.append(f"-id: {p["id"]} - {p["status"]}")

    return "\n".join(output)

def format_territory_state(territory_state):
    output = []
    by_owner = group_by_owner(territory_state)

    output.append(f"--- Territories ---")
    for owner, owned_territories in by_owner.items():
        output.append(f"\n{owner}")
        for t in owned_territories:
            output.append(f" - {t["territory_id"]}")
    
    return "\n".join(output)

def format_units(units):
    output = []
    by_owner = group_by_owner(units)

    output.append(f"--- Units ---")
    for owner, units in by_owner.items():
        output.append(f"\n{owner}")
        for u in units:
            output.append(f" - {u["type"]} - {u["territory_id"]}")
    
    return "\n".join(output)

def group_by_owner(items):
    grouped = defaultdict(list)
    for item in items:
        owner = item.get("owner_id", "None")
        grouped[owner].append(item)
    return grouped

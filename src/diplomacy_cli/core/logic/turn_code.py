from diplomacy_cli.core.logic.schema import Phase, Season

BASE_YEAR = 1901
SEASON_CODES = ("S", "F", "W")
SEASON_NAMES = ("Spring", "Fall", "Winter")
PHASE_CODES = ("M", "R", "A")
PHASE_NAMES = ("Movement", "Retreat", "Adjustment")
INITIAL_TURN_CODE = f"{BASE_YEAR}-{SEASON_CODES[0]}-{PHASE_CODES[0]}"


_TRANSITIONS = {
    (Season.SPRING, Phase.MOVEMENT): (Season.SPRING, Phase.RETREAT, 0),
    (Season.SPRING, Phase.RETREAT): (Season.FALL, Phase.MOVEMENT, 0),
    (Season.FALL, Phase.MOVEMENT): (Season.FALL, Phase.RETREAT, 0),
    (Season.FALL, Phase.RETREAT): (Season.WINTER, Phase.ADJUSTMENT, 0),
    (Season.WINTER, Phase.ADJUSTMENT): (Season.SPRING, Phase.MOVEMENT, 1),
}


def parse_turn_code(turn_code: str) -> tuple[int, Season, Phase]:
    year, season_code, phase_code = turn_code.split("-")
    y_idx = int(year) - BASE_YEAR
    s_idx = SEASON_CODES.index(season_code)
    p_idx = PHASE_CODES.index(phase_code)
    return (y_idx, Season(s_idx), Phase(p_idx))


def format_turn_code(y_idx: int, season: Season, phase: Phase) -> str:
    year = y_idx + BASE_YEAR
    season_code = SEASON_CODES[season.value]
    phase_code = PHASE_CODES[phase.value]
    return f"{year}-{season_code}-{phase_code}"


def advance_turn_tuple(turn_tuple):
    year, season, phase = turn_tuple
    new_season, new_phase, dy = _TRANSITIONS[(season, phase)]
    return (year + dy, new_season, new_phase)


def advance_turn_code(turn_code):
    current_turn = parse_turn_code(turn_code)
    new_year, new_season, new_phase = advance_turn_tuple(current_turn)
    return format_turn_code(new_year, new_season, new_phase)

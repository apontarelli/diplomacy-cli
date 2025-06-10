import unittest

from diplomacy_cli.core.logic.turn_code import (
    Phase,
    Season,
    advance_turn_code,
    advance_turn_tuple,
    format_turn_code,
    parse_turn_code,
)


class TestTurnCode(unittest.TestCase):
    def test_parse_turn_code(self):
        turn_code = "1901-W-A"
        result = parse_turn_code(turn_code)
        expected = (0, Season.WINTER, Phase.ADJUSTMENT)
        self.assertEqual(result, expected)

    def test_format_turn_code(self):
        y_idx, season, phase = (4, Season.FALL, Phase.RETREAT)
        result = format_turn_code(y_idx, season, phase)
        expected = "1905-F-R"
        self.assertEqual(result, expected)

    def test_advance_turn_tuple(self):
        turn_tuple = (34, Season.FALL, Phase.RETREAT)
        result = advance_turn_tuple(turn_tuple)
        expected = (34, Season.WINTER, Phase.ADJUSTMENT)
        self.assertEqual(result, expected)

    def test_advance_turn_code(self):
        turn_code = "1910-F-R"
        result = advance_turn_code(turn_code)
        expected = "1910-W-A"
        self.assertEqual(result, expected)

    def test_advance_turn_code_skip(self):
        turn_code = "1910-F-M"
        result = advance_turn_code(turn_code, True)
        expected = "1910-W-A"
        self.assertEqual(result, expected)


if __name__ == "__main__":
    unittest.main()

import pytest

from diplomacy_cli.core.logic.validator.syntax import normalize_order_string


@pytest.mark.parametrize(
    "raw, expected",
    [
        # Basic normalization: lower, strip, collapse spaces
        ("  BUR  S   PAR - PIC! ", "bur s par - pic"),
        ("Par-Pic", "par - pic"),
        (" MUN    hold", "mun hold"),
        ("stp/sc hold", "stp/sc hold"),
        # Unicode dash normalization
        ("bal c pru – den", "bal c pru - den"),
        ("bal c pru — den", "bal c pru - den"),
        ("bal c pru ‒ den", "bal c pru - den"),
        # Strip unwanted punctuation
        ("par.pic", "parpic"),
        ("bur,s par", "burs par"),
        # Mixed punctuation & whitespace
        ("  Par...   -   Pic??", "par - pic"),
        #
    ],
)
def test_normalize_various_cases(raw, expected):
    assert normalize_order_string(raw) == expected

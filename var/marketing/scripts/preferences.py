"""Child user preferences (theme picker) in the Egypt theme."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from _starapp import CHARLES, prepare


def run(driver):
    prepare(
        driver,
        user=CHARLES,
        theme="egypt",
        path="/user-control-panel/preferences",
        ready="#user-preferences-theme-name",
    )

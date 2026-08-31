"""Family people list as a parent, Aztecs theme in dark mode."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from _starapp import ADMIN, prepare


def run(driver):
    prepare(
        driver,
        user=ADMIN,
        theme="aztecs",
        color_scheme="dark",
        path="/control-panel/people",
        ready=".people-list",
    )

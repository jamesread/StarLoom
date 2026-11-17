"""Privileged Control Panel hub in the Catppuccin supplemental theme."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from _starapp import ADMIN, prepare


def run(driver):
    prepare(
        driver,
        user=ADMIN,
        theme="catppuccin-latte-frappe",
        path="/control-panel",
        ready=".navigation-grid",
    )

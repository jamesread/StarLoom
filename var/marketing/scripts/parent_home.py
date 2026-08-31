"""Privileged parent homepage (family overview) in the Egypt theme."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from _starapp import ADMIN, prepare


def run(driver):
    prepare(driver, user=ADMIN, theme="egypt", ready=".people-card")

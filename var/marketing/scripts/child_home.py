"""Non-privileged child homepage (stars + rewards) in the Aztecs theme."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from _starapp import CHARLES, prepare


def run(driver):
    prepare(driver, user=CHARLES, theme="aztecs", ready=".child-balance")

"""Weekly star chart as a parent, Ancient Greece theme."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from _starapp import ADMIN, prepare


def run(driver):
    prepare(
        driver,
        user=ADMIN,
        theme="ancient-greece",
        path="/family/star-chart/1",
        ready="table.star-chart",
    )

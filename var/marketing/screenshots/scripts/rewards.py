"""Parent rewards catalog and pending redemptions, Space theme in dark mode."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from _starapp import ADMIN, prepare


def run(driver):
    prepare(
        driver,
        user=ADMIN,
        theme="space",
        color_scheme="dark",
        path="/family/rewards",
        ready=".rewards-list",
    )

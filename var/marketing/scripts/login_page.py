"""Logged-out login form in the Space theme."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from _starapp import prepare


def run(driver):
    prepare(driver, theme="space", ready="form.local-login-form")

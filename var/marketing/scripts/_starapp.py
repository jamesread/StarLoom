"""Shared Selenium helpers for StarApp marketing screenshots.

repo-helper screenshot rewrites https:// to http://. The Vite dev server is
HTTPS-only (and session cookies are Secure), so every job starts at about:blank
and these helpers navigate to APP_ORIGIN after ignoring the self-signed cert.
"""

from __future__ import annotations

import time
from urllib.parse import urlparse

from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait

APP_ORIGIN = "https://localhost:5173"
THEME_STORAGE_KEY = "starapp-custom-theme"
COLOR_SCHEME_KEY = "picocrank-theme"

ADMIN = ("admin", "admin")
CHARLES = ("charles", "password")


def ignore_certs(driver):
    try:
        driver.execute_cdp_cmd("Security.setIgnoreCertificateErrors", {"ignore": True})
    except Exception:
        pass


def _maybe_bypass_cert_interstitial(driver):
    source = driver.page_source or ""
    url = driver.current_url or ""
    if not (
        "Your connection is not private" in source
        or "ERR_CERT" in source
        or "privacy-error" in url
    ):
        return
    try:
        driver.find_element(By.TAG_NAME, "body").send_keys("thisisunsafe")
        time.sleep(1.5)
    except Exception:
        pass


def open_path(driver, path="/"):
    ignore_certs(driver)
    if not path.startswith("/"):
        path = "/" + path
    driver.get(APP_ORIGIN + path)
    _maybe_bypass_cert_interstitial(driver)


def wait_css(driver, selector, timeout=15):
    WebDriverWait(driver, timeout).until(
        EC.visibility_of_element_located((By.CSS_SELECTOR, selector))
    )


def is_login_form(driver):
    return any(
        el.is_displayed()
        for el in driver.find_elements(By.CSS_SELECTOR, "form.local-login-form")
    )


def is_app_shell(driver):
    return bool(driver.find_elements(By.CSS_SELECTOR, "#layout"))


def header_username(driver):
    for el in driver.find_elements(By.CSS_SELECTOR, "button.user-info"):
        text = (el.text or "").strip()
        if text:
            return text
    return ""


def apply_theme(driver, theme_name, color_scheme="light"):
    """Persist theme + color scheme on the app origin, then reload."""
    if not (driver.current_url or "").startswith(APP_ORIGIN):
        open_path(driver, "/")
    path = urlparse(driver.current_url).path or "/"
    driver.execute_script(
        """
        const themeKey = arguments[0];
        const theme = arguments[1] || '';
        const schemeKey = arguments[2];
        const scheme = arguments[3] || 'light';
        if (theme) {
          localStorage.setItem(themeKey, theme);
        } else {
          localStorage.removeItem(themeKey);
        }
        localStorage.setItem(schemeKey, scheme);
        """,
        THEME_STORAGE_KEY,
        theme_name or "",
        COLOR_SCHEME_KEY,
        color_scheme,
    )
    open_path(driver, path)
    if theme_name == "ancient-greece":
        try:
            WebDriverWait(driver, 8).until(
                lambda d: d.execute_script(
                    "return document.fonts && document.fonts.check(\"1em 'GreeKish'\")"
                )
            )
        except Exception:
            time.sleep(1.0)
    time.sleep(0.35)


def ensure_logged_out(driver):
    open_path(driver, "/")
    if is_login_form(driver):
        return
    try:
        driver.execute_script(
            """
            return fetch('/api/starapp.api.v1.StarAppService/Logout', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
              },
              credentials: 'include',
              body: '{}'
            });
            """
        )
    except Exception:
        pass
    try:
        driver.delete_cookie("starapp-sid")
    except Exception:
        pass
    open_path(driver, "/")
    WebDriverWait(driver, 15).until(
        EC.visibility_of_element_located((By.CSS_SELECTOR, "form.local-login-form"))
    )


def login(driver, username, password):
    open_path(driver, "/")
    if is_app_shell(driver) and header_username(driver).lower() == username.lower():
        return
    if is_app_shell(driver):
        ensure_logged_out(driver)
    else:
        wait_css(driver, "#username")
    user = driver.find_element(By.ID, "username")
    pwd = driver.find_element(By.ID, "password")
    user.clear()
    user.send_keys(username)
    pwd.clear()
    pwd.send_keys(password)
    driver.find_element(By.CSS_SELECTOR, "form.local-login-form button[type='submit']").click()
    WebDriverWait(driver, 15).until(
        EC.presence_of_element_located((By.CSS_SELECTOR, "#layout"))
    )
    time.sleep(0.4)


def prepare(
    driver,
    *,
    user=None,
    theme="",
    color_scheme="light",
    path="/",
    ready=None,
):
    """Open the app, sign in if needed, apply a theme, then wait for `ready`."""
    if user is None:
        ensure_logged_out(driver)
    else:
        login(driver, user[0], user[1])
    apply_theme(driver, theme, color_scheme)
    current = urlparse(driver.current_url).path or "/"
    if path != current:
        open_path(driver, path)
    if ready:
        wait_css(driver, ready)
    time.sleep(0.2)

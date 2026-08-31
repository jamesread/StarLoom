"""Force repo-helper to skip a stale webdriver_manager cache.

repo-helper imports ChromeDriverManager when that package is installed. The
cached driver is often far behind Fedora Chromium; raising ImportError here
makes repo-helper fall back to Selenium + the system `chromedriver` on PATH.
"""

raise ImportError("use the system chromedriver (see var/marketing/chromedriver-shim)")

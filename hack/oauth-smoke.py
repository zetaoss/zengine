"""Exercise a temporary wiki's owner-only OAuth token through the Action API."""

import http.cookiejar
import json
import time
import urllib.error
import urllib.parse
import urllib.request

BASE = "http://127.0.0.1:8080/w/"


def new_client():
    return urllib.request.build_opener(
        urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar())
    )


client = new_client()


def request(path, data=None, headers=None):
    req = urllib.request.Request(BASE + path, data=data, headers=headers or {})
    with client.open(req, timeout=10) as response:
        return json.load(response)


def api(params, *, post=False, headers=None):
    data = urllib.parse.urlencode({**params, "format": "json"}).encode()
    path = "api.php" if post else "api.php?" + data.decode()
    return request(path, data if post else None, headers)


for attempt in range(30):
    try:
        login_token = api({"action": "query", "meta": "tokens", "type": "login"})[
            "query"
        ]["tokens"]["logintoken"]
        break
    except urllib.error.URLError:
        if attempt == 29:
            raise
        time.sleep(0.2)

result = api(
    {
        "action": "clientlogin",
        "logintoken": login_token,
        "username": "TestAdmin",
        "password": "TemporaryPass123",
        "loginreturnurl": BASE,
    },
    post=True,
)
assert result["clientlogin"]["status"] == "PASS", result

registration = {
    "name": "mwtool-test",
    "description": "temporary test",
    "email": "test@example.com",
    "is_confidential": True,
    "owner_only": True,
    "grant_types": ["client_credentials"],
    "scopes": ["basic", "editpage"],
}
result = request(
    "rest.php/oauth2/client",
    json.dumps(registration).encode(),
    {"Content-Type": "application/json"},
)
token = result.get("access_token")
assert token, "No owner-only token returned"

# Use a fresh cookie jar to prove that the bearer token, not the login cookie, authenticates.
client = new_client()
result = api({"action": "query", "meta": "userinfo", "assert": "user"})
assert result.get("error", {}).get("code") == "assertuserfailed", result
headers = {"Authorization": "Bearer " + token}
result = api({"action": "query", "meta": "userinfo", "assert": "user"}, headers=headers)
assert result["query"]["userinfo"]["name"] == "TestAdmin", result
csrf = api(
    {"action": "query", "meta": "tokens", "type": "csrf", "assert": "user"},
    headers=headers,
)["query"]["tokens"]["csrftoken"]
result = api(
    {"action": "edit", "title": "OAuth new page", "text": "test", "token": csrf, "assert": "user"},
    post=True,
    headers=headers,
)
assert result.get("error", {}).get("code") == "cantcreate", result
result = api(
    {"action": "edit", "title": "Main Page", "text": "test content", "token": csrf, "assert": "user"},
    post=True,
    headers=headers,
)
assert result.get("edit", {}).get("result") == "Success", result
print("OAuth owner-only token issued; authenticated API query and edit succeeded")

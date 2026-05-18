# ── 01-auth.hurl ──────────────────────────────────────────────────────────────
# Tests: POST /auth/token for all three roles, and invalid role → 400.

# Admin token
POST http://localhost:1323/auth/token
Content-Type: application/json
{
  "name": "hurl-admin-01",
  "role": "admin"
}

HTTP 200
[Captures]
admin_token: jsonpath "$.token"
[Asserts]
jsonpath "$.token" isString
jsonpath "$.principal.name" == "hurl-admin-01"
jsonpath "$.principal.role" == "admin"
jsonpath "$.principal.id" isInteger


# Enduser token
POST http://localhost:1323/auth/token
Content-Type: application/json
{
  "name": "hurl-user-01",
  "role": "enduser"
}

HTTP 200
[Asserts]
jsonpath "$.token" isString
jsonpath "$.principal.role" == "enduser"


# Drone token
POST http://localhost:1323/auth/token
Content-Type: application/json
{
  "name": "hurl-drone-01",
  "role": "drone"
}

HTTP 200
[Asserts]
jsonpath "$.token" isString
jsonpath "$.principal.role" == "drone"


# Invalid role → 400
POST http://localhost:1323/auth/token
Content-Type: application/json
{
  "name": "hurl-bad-01",
  "role": "superuser"
}

HTTP 400


# Missing name → 400
POST http://localhost:1323/auth/token
Content-Type: application/json
{
  "role": "admin"
}

HTTP 400

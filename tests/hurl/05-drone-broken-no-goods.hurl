# ── 05-drone-broken-no-goods.hurl ─────────────────────────────────────────────
# Tests: drone self-reports broken while not carrying any goods.
#   No handoff job should be created.

POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-drone-05", "role": "drone"}

HTTP 200
[Captures]
drone_token: jsonpath "$.token"


# Drone marks itself broken (not carrying any order)
POST http://localhost:1323/drone/self/broken
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.drone.status" == "BROKEN"


# No HANDOFF_PICKUP jobs exist
GET http://localhost:1323/drone/jobs?type=HANDOFF_PICKUP
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.jobs" count == 0


# GET /drone/self/order — no assigned order
GET http://localhost:1323/drone/self/order
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.drone.status" == "BROKEN"
jsonpath "$.order" == null

# ── 02-enduser.hurl ───────────────────────────────────────────────────────────
# Tests: POST /orders, GET /orders/{id} (with timeline), POST /orders/{id}/withdraw.

POST http://localhost:1323/auth/token
Content-Type: application/json
{
  "name": "hurl-user-02",
  "role": "enduser"
}

HTTP 200
[Captures]
user_token: jsonpath "$.token"
user_id: jsonpath "$.principal.id"


# Submit order → 201
POST http://localhost:1323/orders
Authorization: Bearer {{user_token}}
Content-Type: application/json
{
  "origin":      {"lat": 30.0444, "lng": 31.2357},
  "destination": {"lat": 30.0626, "lng": 31.2497}
}

HTTP 201
[Captures]
order_id: jsonpath "$.id"
[Asserts]
jsonpath "$.status" == "READY_FOR_PICKUP"
jsonpath "$.origin_lat" == 30.0444
jsonpath "$.origin_lng" == 31.2357
jsonpath "$.dest_lat" == 30.0626
jsonpath "$.dest_lng" == 31.2497
jsonpath "$.enduser_principal_id" == {{user_id}}


# Get order — includes timeline
GET http://localhost:1323/orders/{{order_id}}
Authorization: Bearer {{user_token}}

HTTP 200
[Asserts]
jsonpath "$.id" == {{order_id}}
jsonpath "$.status" == "READY_FOR_PICKUP"
jsonpath "$.timeline" isCollection
jsonpath "$.timeline" count >= 1


# Withdraw order
POST http://localhost:1323/orders/{{order_id}}/withdraw
Authorization: Bearer {{user_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "WITHDRAWN"
jsonpath "$.withdrawn_at" isString


# Cannot withdraw the same order twice → 409
POST http://localhost:1323/orders/{{order_id}}/withdraw
Authorization: Bearer {{user_token}}

HTTP 409


# Verify final status via GET
GET http://localhost:1323/orders/{{order_id}}
Authorization: Bearer {{user_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "WITHDRAWN"

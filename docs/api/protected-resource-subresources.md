# Protected Resource Subresources API

> **Implemented API — Feature 035.** This guide describes the protected-resource subresources and full-service update behavior implemented by the admin server. The canonical API definition is [`api/admin/openapi.yaml`](../../api/admin/openapi.yaml).

Protected resources are absolute URIs that map an RFC 8693 token-exchange request to a third-party service. The API normalizes a URI by trimming trailing path slashes before storing and matching it. A normalized URI may belong to only one service globally.

All endpoints below use the existing administrator authentication and authorization applicable to `/api/services`. Replace `SERVICE_ID` with the service UUID. Examples use documentation-only URIs and contain no credentials.

## Read the resource set

```
GET /api/services/{SERVICE_ID}/protected-resources
```

A successful response returns the normalized set and the service's strong `ETag`. Retain that ETag when a later full-service update needs to replace the whole set.

```http
GET /api/services/8e5aa5aa-8ec3-4c45-842b-55050a359bb4/protected-resources HTTP/1.1
```

```http
HTTP/1.1 200 OK
ETag: "7"
Content-Type: application/json

{
  "protected_resources": [
    "https://api.example.com/v1",
    "https://payments.example.net"
  ]
}
```

A missing service returns `404`.

## Add one resource

### Collection POST

Use `POST` when the URI naturally belongs in a request body:

```http
POST /api/services/8e5aa5aa-8ec3-4c45-842b-55050a359bb4/protected-resources HTTP/1.1
Content-Type: application/json

{
  "resource_uri": "https://api.example.com/v2/"
}
```

The server validates and normalizes the URI, yielding `https://api.example.com/v2`. A new claim returns `201 Created` and a new `ETag`; replaying a normalized URI already owned by that same service returns `200 OK` without a duplicate. If another service owns the normalized URI, the response is `409 Conflict`. Malformed, relative, empty, or whitespace-only values return `400 Bad Request` before state changes.

### Member-addressed PUT

`PUT` is the retained idempotent add operation. It has no request body; the member URI is the final path segment:

```sh
curl --path-as-is -X PUT \
  -H "X-Remote-User: admin@example.com" \
  'http://localhost:14000/api/services/8e5aa5aa-8ec3-4c45-842b-55050a359bb4/protected-resources/https%3A%2F%2Fapi.example.com%2Fv2'
```

Its status and ownership behavior match collection `POST`: `201` for a new URI, `200` for an idempotent replay, `409` if another service owns the normalized URI, and `400` for an invalid URI. Successful responses include the resulting set and an `ETag`.

## Address a member URI safely

A member URI is not a hierarchy of API paths. It must occupy **one fully RFC 3986 percent-encoded path segment**. Encode every reserved character in the URI, including `:` as `%3A`, `/` as `%2F`, `?` as `%3F`, and `#` as `%23`. For example, this source URI:

```
https://api.example.com/v2?region=eu#status
```

is addressed as:

```
https%3A%2F%2Fapi.example.com%2Fv2%3Fregion%3Deu%23status
```

The server obtains the escaped segment, confirms it is exactly one segment, then percent-decodes it exactly once before validation, normalization, and matching. Do not submit a raw URI containing `/`, pre-decode the value, or double-encode it. If the URI itself contains a percent sign, encode that percent sign as `%25` in the path segment. Gateways in front of the API must preserve encoded slashes rather than decoding or rejecting `%2F`.

## Remove one resource

```
DELETE /api/services/{SERVICE_ID}/protected-resources/{ENCODED_RESOURCE_URI}
```

```http
DELETE /api/services/8e5aa5aa-8ec3-4c45-842b-55050a359bb4/protected-resources/https%3A%2F%2Fapi.example.com%2Fv2 HTTP/1.1
```

A successful removal returns `200 OK` with the removed normalized URI, the resulting set, and the current strong `ETag`. Removing a URI the service does not own, or targeting a missing service, returns `404`; an invalid encoded member URI returns `400`. Removing a URI stops new token-exchange resolution for that URI; it does not revoke already-issued downstream tokens.

## Rename one resource

`PATCH` uses the encoded path member as the source URI and `to` as the target URI. The operation is atomic; clients do not need a delete-then-add round trip.

```http
PATCH /api/services/8e5aa5aa-8ec3-4c45-842b-55050a359bb4/protected-resources/https%3A%2F%2Fapi.example.com%2Fv1 HTTP/1.1
Content-Type: application/json

{
  "to": "https://api.example.com/v2"
}
```

A successful rename returns `200 OK` with the affected normalized URI, the resulting set, and an `ETag`. Renaming a URI to itself succeeds as a no-op. A missing source or service returns `404`; an invalid source or target returns `400`; a target already owned by any service, including the current service, returns `409`.

## Full-service update and ETag retry

The existing full-service endpoint retains whole-set replacement, but it is intentionally distinct from the single-resource operations:

```
PUT /api/services/{SERVICE_ID}
```

- If `protected_resources` is omitted or `null`, the API preserves the current resource set. No `If-Match` is required solely because of this update.
- If `protected_resources` is present, including an empty array, it authoritatively replaces the set. The request must carry the current strong ETag in `If-Match`.
- An empty array is an explicit request to clear the set, not an omission.

For a replacement, first read the current resource set (or service) and use its ETag. Include every other field required by the existing full-service update contract; the fragment below shows only the resource-set behavior:

```http
PUT /api/services/8e5aa5aa-8ec3-4c45-842b-55050a359bb4 HTTP/1.1
If-Match: "7"
Content-Type: application/json

{
  "protected_resources": [
    "https://api.example.com/v2",
    "https://payments.example.net"
  ]
}
```

If a concurrent add, remove, rename, or other service mutation advances the version, this request receives `412 Precondition Failed` and does not modify the set. Retry by re-reading the current service or protected-resource collection, reconciling the desired authoritative set with the returned state, and sending a new replacement with the fresh ETag. Do not blindly resend the old set, because doing so may intentionally discard a change that the precondition protected.

A replacement that omits `If-Match` returns `428 Precondition Required`. Conflicts with a URI claimed by another service return `409 Conflict`. Successful replacements return `200 OK` and a new `ETag`.

For an update that does not intend to change protected resources, omit the field:

```http
PUT /api/services/8e5aa5aa-8ec3-4c45-842b-55050a359bb4 HTTP/1.1
Content-Type: application/json

{
  "display_name": "Example Payments Service"
}
```

As with every full-service update, include the fields required by the canonical service schema. The omitted `protected_resources` field leaves the resource set untouched.

## Status summary

| Operation | Success | Important failures |
|---|---|---|
| `GET` collection | `200` + set and `ETag` | `404` service missing |
| `POST` collection / member `PUT` | `201` new, `200` idempotent | `400` invalid URI, `404` service missing, `409` globally claimed |
| `PATCH` member | `200` + result and `ETag` | `400` invalid URI, `404` source or service missing, `409` claimed target |
| `DELETE` member | `200` + result and `ETag` | `400` invalid URI, `404` source or service missing |
| Full-service replacement | `200` + `ETag` | `409` globally claimed URI, `412` stale ETag, `428` missing `If-Match` |

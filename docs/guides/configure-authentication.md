---
title: "Configure authentication"
description: "Set up the trusted reverse proxy that authenticates users and injects the principal header, and optionally verify a signed JWT and extract a richer profile."
---

# Configure authentication

The broker does not authenticate users. It delegates that job to a trusted reverse proxy in
front of it — oauth2-proxy, an nginx `auth_request`, an API gateway, or a service mesh. The
proxy authenticates the request against your identity provider and injects the user's
identity in a request header. The broker reads that header as the **principal** and trusts
it only because it comes from a source you control.

This page shows you how to configure that boundary: the principal header on both server
ports, the trust requirement that keeps it safe, and an optional JWT pre-authentication mode
that verifies a signed token and extracts a profile. For where this fits in the overall
system, see [architecture](/docs/concepts/architecture).

## What you'll need

- A reverse proxy or gateway that already authenticates your users and can set (and strip)
  request headers. Keep using your existing identity provider — Keycloak, Okta, Auth0, an
  internal OIDC provider — for human login.
- Network access to place that proxy in front of the broker's two ports: the end-user API on
  **8000** and the admin API on **14000**.
- Access to the broker's YAML configuration (see the [configuration reference](/docs/configuration)).
- For JWT pre-authentication: a JWKS endpoint your proxy's signed tokens can be verified
  against, or a trusted mesh that injects unsigned claims.

## Reverse-proxy pre-authentication

This is the baseline mode and the one every deployment uses. The proxy sets a header whose
value is the principal identifier — an email, a username, or an opaque ID. The broker takes
that value as the authenticated user for the request.

### Set the principal header

Configure the header name under `authentication.preauth.principal_header_name`. Each server
port has its own block, so set it on both `enduser` and `admin`. The default is
`X-Remote-User`.

```yaml
server:
  enduser:
    port: 8000
    bind: "0.0.0.0"
    public_url: "https://broker.example.com"
    authentication:
      preauth:
        principal_header_name: "X-Remote-User"
  admin:
    port: 14000
    bind: "0.0.0.0"
    public_url: "https://broker.example.com:14000"
    authentication:
      preauth:
        principal_header_name: "X-Remote-User"
```

Point your proxy at the matching header name. If your proxy emits `X-Forwarded-User` or
`X-Auth-User`, either rename it at the proxy or set `principal_header_name` to match — the
two must agree.

### Establish the trust boundary

The principal header is an assertion of identity. It is only safe if a request can never
reach the broker carrying a header the proxy did not set. Two rules make that true:

- **Only the proxy may reach the broker ports.** Bind the broker to an internal network and
  route all traffic through the proxy. A client that can open a direct connection to port
  8000 or 14000 can assert any principal.
- **The proxy must strip client-supplied copies.** Before the proxy adds its own trusted
  value, it must remove any inbound `X-Remote-User` (or your configured name) from the
  incoming request. Otherwise a caller can pre-set the header and impersonate another user.

:::warning
Treat the principal header exactly as you would a bearer token: the proxy is the only party
allowed to produce it, and any copy arriving from the outside must be dropped before the
request is authenticated. Exposing the broker ports directly, or forgetting to strip the
inbound header, defeats the entire delegation model.
:::

### Enforce admin privilege at the proxy

The admin API on port 14000 accepts the same principal header. The broker does **not**
distinguish administrators from ordinary users — administrative privilege is enforced at the
proxy, before requests reach the admin port. Restrict the admin route to your operators
using the proxy's own access controls (group membership, an allowlist, a separate
authentication policy, or a dedicated ingress). Keep the admin port internal and never expose
it alongside the public end-user port. The full admin surface is described in the
[API reference](/docs/reference/api).

## JWT pre-authentication (optional)

If your proxy or mesh issues a JWT rather than a plain header value, you can enable JWT
pre-authentication. The broker reads the token from a header, optionally verifies its
signature, and extracts the principal and a profile (display name, email, picture) using CEL
expressions. This is how the consent UI shows a user's name and avatar instead of a bare ID.

Configure it under `server.enduser.authentication.jwt` (and `server.admin` if the admin
proxy issues JWTs too). Keep the `preauth` block as well — the server config still requires
`principal_header_name`, but once `jwt` is enabled the middleware rejects requests missing
`header_name` with `401 Unauthorized` instead of falling back to the plain header.

### Verify against a JWKS (recommended)

Use this when your proxy issues signed JWTs and you want cryptographic verification. Set
`verification: jwks` (the default when a `jwks_uri` is present) and pin the expected audience
and issuer so tokens minted for other services are rejected.

```yaml
server:
  enduser:
    port: 8000
    authentication:
      preauth:
        principal_header_name: "X-Remote-User"   # still configured, but not used as a runtime fallback
      jwt:
        header_name: "Authorization"             # strips the "Bearer " prefix
        verification: "jwks"
        jwks_uri: "https://auth.example.com/.well-known/jwks.json"
        expected_audience: "agentic-identity-broker"
        expected_issuer: "https://auth.example.com"
        claim_extraction:
          principal_expression: "claims.sub"
          display_name_expression: "claims.name"
          email_expression: "claims.email"
          picture_url_expression: "claims.picture"
```

### Trust a mesh without verifying (`verification: none`)

Use this only inside a network where you already trust that JWT injection is controlled — for
example a mutual-TLS service mesh (Istio, Linkerd) or a trusted sidecar that has already
authenticated the user. With `verification: none`, the broker parses the claims without
checking a signature.

```yaml
server:
  enduser:
    port: 8000
    authentication:
      preauth:
        principal_header_name: "X-Remote-User"
      jwt:
        header_name: "X-JWT-Claims"
        verification: "none"
        claim_extraction:
          principal_expression: "claims.sub"
          display_name_expression: "claims.preferred_username"
          email_expression: "claims.email"
```

Do not set `jwks_uri` when `verification` is `none` — the two are mutually exclusive and the
broker refuses to start if both are present.

### Extract the principal and profile with CEL

Each `claim_extraction` expression is a CEL expression evaluated against the token's `claims`
object. Only `principal_expression` is required; it must resolve to the string used as the
principal. The profile expressions are optional and populate the values the consent UI shows:

| Expression | Populates | Typical claim |
|---|---|---|
| `principal_expression` | The principal identifier (required) | `claims.sub` |
| `display_name_expression` | The user's display name | `claims.name` or `claims.preferred_username` |
| `email_expression` | The user's email | `claims.email` |
| `picture_url_expression` | The user's avatar URL | `claims.picture` |

Adjust each expression to match the claim names your identity provider emits.

## Troubleshooting

- **Every request returns 401.** The principal is missing or empty. Confirm the proxy is
  setting the configured header on requests to the broker, and that the header name in the
  proxy matches `principal_header_name` exactly. With JWT pre-authentication, check that the
  token is present in `header_name` and that `principal_expression` resolves to a non-empty
  value.
- **Requests succeed but the wrong user is authenticated, or authentication succeeds from
  outside the proxy.** The header is being trusted from an untrusted source. Verify the
  broker ports are reachable only through the proxy, and that the proxy strips any
  client-supplied copy of the header before it sets its own.
- **The broker refuses to start with a JWT error.** You have set both `verification: none`
  and `jwks_uri`; they are mutually exclusive. Remove `jwks_uri` for an unsigned mesh, or set
  `verification: jwks` to verify.
- **The consent UI shows a bare ID instead of a name or avatar.** The profile expressions did
  not resolve. Confirm the JWT actually carries those claims and that
  `display_name_expression`, `email_expression`, and `picture_url_expression` reference the
  correct claim names.

## Related

- [Architecture](/docs/concepts/architecture) — where the proxy sits and how a request flows.
- [Configuration reference](/docs/configuration) — every configuration key in detail.
- [API reference](/docs/reference/api) — the endpoints protected by this authentication.

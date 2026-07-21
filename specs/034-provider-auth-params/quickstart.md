# Quickstart: Validate Provider Authorization Parameters

## Prerequisites

- A local broker development environment with the admin server and end-user server available.
- An administrator principal accepted by the local reverse-proxy configuration.
- A third-party OAuth2 provider or a test upstream authorization endpoint.

See the [data model](data-model.md) and [admin contract](contracts/admin-service-authorization-params.md) for the field and validation rules.

## 1. Create a Zalando Platform service

Create a normal third-party OAuth2 service with static provider configuration:

```json
{
  "display_name": "Zalando Platform",
  "client_id": "configured-client-id",
  "client_secret": "configured-client-secret",
  "issuer_uri": "https://provider.example.com",
  "discovery": { "enable_discovery": false },
  "endpoints": {
    "authorize_endpoint": "https://provider.example.com/oauth/authorize",
    "token_endpoint": "https://provider.example.com/oauth/token"
  },
  "scopes": [
    { "scope_value": "profile", "description": "Read profile" }
  ],
  "authorization_params": {
    "business_partner_id": "12345"
  }
}
```

**Expected result**: The create response returns `authorization_params.business_partner_id` as `12345`, while the client secret remains redacted.

## 2. Verify administrative reads and update preservation

1. Retrieve the service and list services; both responses include the configured map.
2. Update another service property while omitting `authorization_params`.
3. Retrieve the service again.

**Expected result**: The stored `business_partner_id` remains `12345`. An explicit empty object is the intentional way to clear the map.

## 3. Start authorization

Open the broker's existing authorization-start endpoint with its normal same-origin post-session `redirect_uri`.

**Expected result**: The redirect location targets the configured upstream authorization endpoint and includes:

- `business_partner_id=12345`
- broker-generated `client_id`, `redirect_uri`, `response_type=code`, `scope`, and `state`
- PKCE `code_challenge` and `code_challenge_method` when the normal flow uses PKCE

## 4. Verify query isolation

Repeat the authorization-start request with `business_partner_id=untrusted` in the browser query string.

**Expected result**: The upstream redirect still contains only `business_partner_id=12345`; `untrusted` is not forwarded.

## 5. Verify validation

Attempt create or update requests with each of the following:

- blank or whitespace-only parameter names or values
- any broker-owned key from the [admin contract](contracts/admin-service-authorization-params.md#validation)

**Expected result**: Each request receives the existing validation-error response and does not alter stored provider configuration.

## Automated verification

Run focused unit and E2E tests while developing, then use the repository verification commands:

```bash
just check
just test
just test-e2e-backend
just test-integration
just test-integration-infra
```

The PostgreSQL suite requires Docker or Podman. The standard full verification command may be used once the focused suites pass.

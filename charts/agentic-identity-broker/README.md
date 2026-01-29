# Agentic Identity Broker Helm Chart

Official Helm chart for deploying the Agentic Identity Broker on Kubernetes.

## Features

- 🚀 **Simple deployment** with sensible defaults
- 🔐 **Flexible storage** options: in-memory, external PostgreSQL, or Zalando PostgreSQL Operator
- 🔄 **Automatic migrations** via Helm hooks with separate credentials
- 🌐 **Dual Ingress** support for end-user and admin APIs
- 📊 **Production-ready** with Pod Security Standards, resource limits, and health checks
- 📦 **Static manifests** generation via `helm template`
- ⚙️ **Highly configurable** with comprehensive values.yaml

## Prerequisites

- Kubernetes 1.25+
- Helm 3.x
- (Optional) Zalando PostgreSQL Operator for managed database provisioning

## Quick Start

### Install with In-Memory Storage

For evaluation and testing:

```bash
helm install broker ./charts/agentic-identity-broker
```

⚠️ **Warning**: In-memory storage is ephemeral. Data is lost on pod restart. Use PostgreSQL for production.

### Install with External PostgreSQL

For production deployments with an existing PostgreSQL database:

1. Create Kubernetes Secrets for database credentials:

```bash
# Migration user (schema permissions)
kubectl create secret generic broker-db-migration \
  --from-literal=username=broker-migration \
  --from-literal=password=<migration-password>

# Broker user (data access)
kubectl create secret generic broker-db \
  --from-literal=username=broker \
  --from-literal=password=<broker-password>
```

2. Install the chart:

```bash
helm install broker ./charts/agentic-identity-broker \
  --set storage.type=postgres \
  --set postgresql.external.enabled=true \
  --set postgresql.external.host=postgres.database.svc.cluster.local \
  --set postgresql.external.database=broker \
  --set postgresql.external.migrationSecretName=broker-db-migration \
  --set postgresql.external.brokerSecretName=broker-db
```

### Install with Zalando PostgreSQL Operator

For automatic PostgreSQL provisioning:

```bash
helm install broker ./charts/agentic-identity-broker \
  --set storage.type=postgres \
  --set postgresql.operator.enabled=true \
  --set postgresql.operator.teamId=my-team
```

The operator will:
- Create a PostgreSQL cluster with 2 instances
- Create migration and broker users
- Generate Kubernetes Secrets automatically
- Configure the broker to use the operator-managed database

## Configuration

See [values.yaml](values.yaml) for the complete list of configuration options.

### Key Configuration Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of broker pods | `1` |
| `image.repository` | Broker container image repository | `agentic-identity-broker` |
| `image.tag` | Broker image tag | Chart appVersion |
| `migration.image.repository` | Migration container image repository | `agentic-identity-broker-migrate` |
| `migration.image.tag` | Migration image tag | Chart appVersion |
| `storage.type` | Storage backend (`memory` or `postgres`) | `memory` |
| `ingress.enduser.enabled` | Enable Ingress for end-user API | `false` |
| `ingress.admin.enabled` | Enable Ingress for admin API | `false` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |

### Custom Values File

Create a `values-production.yaml` file:

```yaml
replicaCount: 3

storage:
  type: postgres

postgresql:
  operator:
    enabled: true
    teamId: production
    numberOfInstances: 3
    volume:
      size: 50Gi

ingress:
  enduser:
    enabled: true
    className: nginx
    hosts:
      - host: broker.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: broker-tls
        hosts:
          - broker.example.com

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

Install with custom values:

```bash
helm install broker ./charts/agentic-identity-broker -f values-production.yaml
```

## Ingress Configuration

### Nginx Ingress Controller

```yaml
ingress:
  enduser:
    enabled: true
    className: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
    hosts:
      - host: broker.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: broker-tls
        hosts:
          - broker.example.com
```

### Skipper Ingress Controller

```yaml
ingress:
  enduser:
    enabled: true
    className: skipper
    annotations:
      zalando.org/skipper-filter: ratelimit(50, "1m")
      zalando.org/skipper-predicate: 'Path("/")'
    hosts:
      - host: broker.example.com
        paths:
          - path: /
            pathType: ImplementationSpecific
```

## Database Migration

### How Migrations Work

1. Migrations run as a Kubernetes Job with Helm hook (`pre-install`, `pre-upgrade`)
2. Migration Job uses a **separate Docker image** (`agentic-identity-broker-migrate`) containing only migration tooling
3. Migration Job uses a separate ServiceAccount and user credentials (least privilege)
4. Broker Deployment uses the main image and a different user with appropriate runtime permissions
5. Migration Job completes before broker pods start

### Security: Separate Migration Image

For security hardening, the migration Job uses a dedicated Docker image:

| Image | Contains | Used By | Security Posture |
|-------|----------|---------|------------------|
| `agentic-identity-broker` | Broker binary, frontend assets | Broker Deployment | Minimal - no SQL migration capabilities |
| `agentic-identity-broker-migrate` | golang-migrate tool, migration files | Migration Job only | Schema permissions - isolated to Job |

**Benefits**:
- **Defense in Depth**: Runtime broker containers cannot execute arbitrary SQL migrations even if compromised
- **Least Privilege**: Migration tooling only present in short-lived Job pods
- **Clear Security Boundaries**: Explicit separation between runtime and schema modification capabilities

Both images share the same version tag to ensure consistency between migrations and application code.

### Migration Credentials

**External PostgreSQL**:
- Set `postgresql.external.migrationSecretName` to your migration user secret
- Set `postgresql.external.brokerSecretName` to your broker user secret
- Migration user needs: `CREATE`, `ALTER`, `DROP` permissions for schema changes
- Broker user needs: `SELECT`, `INSERT`, `UPDATE`, `DELETE` permissions for data access

**Zalando PostgreSQL Operator**:
- Secrets are created automatically by the operator
- Secret names follow the pattern: `{username}.{teamId}-{instanceName}.credentials.postgresql.acid.zalan.do`
- Migration user gets: `superuser` and `createdb` role attributes
- Broker user gets: `login` attribute + database ownership (grants full table access)

## Static Manifest Generation

Generate Kubernetes manifests without installing:

```bash
# Generate with default values
helm template broker ./charts/agentic-identity-broker > manifests.yaml

# Generate with custom values
helm template broker ./charts/agentic-identity-broker \
  --set storage.type=postgres \
  --set postgresql.operator.enabled=true \
  --set postgresql.operator.teamId=my-team \
  > manifests.yaml

# Apply to cluster
kubectl apply -f manifests.yaml
```

## Upgrading

```bash
# Upgrade to new chart version or values
helm upgrade broker ./charts/agentic-identity-broker -f values-production.yaml

# Migrations run automatically before broker pods are updated
```

## Uninstalling

```bash
helm uninstall broker
```

⚠️ **Note**: If using Zalando PostgreSQL Operator, the PostgreSQL CR is NOT deleted automatically to prevent data loss. Delete manually if desired:

```bash
kubectl delete postgresql <teamId>-broker
```

## Troubleshooting

### Migration Job Failed

Check migration logs:

```bash
kubectl logs job/broker-migrate
```

Common issues:
- **Database not reachable**: Check network policies, service DNS resolution
- **Invalid credentials**: Verify secret values and key names
- **Migration files missing**: Ensure Docker image includes `/app/migrations/`

### Broker Pod CrashLoopBackOff

Check broker logs:

```bash
kubectl logs deployment/broker
```

Common issues:
- **Database connection failed**: Verify broker secret and database host
- **Configuration error**: Check ConfigMap values
- **Missing JWE signing key**: Ensure secret `broker-jwe-key` exists

### Zalando Operator Secrets Not Created

Check PostgreSQL CR status:

```bash
kubectl describe postgresql <teamId>-broker
```

The operator creates secrets after the PostgreSQL cluster is ready (1-2 minutes).

## Security

### Pod Security Standards

The chart enforces Kubernetes Pod Security Standards (restricted):

- `runAsNonRoot: true` - Runs as non-root user (UID 1000)
- `readOnlyRootFilesystem: true` - Filesystem is read-only
- `allowPrivilegeEscalation: false` - No privilege escalation
- `capabilities.drop: [ALL]` - All Linux capabilities dropped
- `seccompProfile: RuntimeDefault` - Seccomp profile applied

### Database Credential Separation

- **Migration Job**: Uses dedicated ServiceAccount with separate database credentials
- **Migration user**: Schema modification permissions (varies by PostgreSQL setup)
- **Broker user**: Runtime data access permissions (varies by PostgreSQL setup)
- **Separate Secrets**: Different Kubernetes Secrets for each user

## Examples

### Minimal Production Setup

```yaml
storage:
  type: postgres

postgresql:
  operator:
    enabled: true
    teamId: production

ingress:
  enduser:
    enabled: true
    hosts:
      - host: broker.example.com
        paths:
          - path: /
            pathType: Prefix
```

### High Availability Setup

```yaml
replicaCount: 3

storage:
  type: postgres

postgresql:
  operator:
    enabled: true
    teamId: production
    numberOfInstances: 3

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10

podDisruptionBudget:
  enabled: true
  minAvailable: 2
```

## License

[License information]

## Maintainers

- Agentic Identity Broker Team

## Links

- [Project Repository](https://github.com/yourusername/agentic-identity-broker)
- [Documentation](https://github.com/yourusername/agentic-identity-broker/tree/main/docs)
- [Issue Tracker](https://github.com/yourusername/agentic-identity-broker/issues)

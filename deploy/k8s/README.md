# Kubernetes deployment

Plain manifests, no Helm. The defaults deploy a single replica backed by SQLite
on a PersistentVolumeClaim, which is all most installs need.

## Apply

```bash
# 1. Pin a signing key (optional — one is generated on first start otherwise)
kubectl create secret generic openproxy-secrets \
  --from-literal=JWT_SIGN_KEY="$(openssl rand -hex 32)"

# 2. Apply everything except the ingress
kubectl apply -f pvc.yaml -f deployment.yaml -f service.yaml

# or, with kustomize:
kubectl apply -k .
```

Then create the administrator account through the UI:

```bash
kubectl port-forward svc/openproxy 8081:8081
# open http://localhost:8081 — a fresh database redirects to /setup
```

## Choosing an image tag

`kustomization.yaml` pins `ghcr.io/fdddf/openproxy:latest`. For reproducible
rollouts point it at a released tag instead:

```bash
kustomize edit set image ghcr.io/fdddf/openproxy=ghcr.io/fdddf/openproxy:v1.2.3
```

## Exposing it

`ingress.yaml` is not applied by default because it needs a real hostname. The
proxy terminates no TLS of its own, so it must sit behind an ingress or another
TLS terminator before it is reachable outside the cluster. The annotations there
turn off response buffering, which streaming completions require.

## Scaling past one replica

SQLite on a ReadWriteOnce volume allows exactly one writer, so the Deployment is
fixed at `replicas: 1` with the `Recreate` strategy. To run more:

1. Point the Deployment at Postgres — set `DB_DRIVER` to `postgres` and
   uncomment the `DB_*` block in `deployment.yaml`.
2. Add `DB_PASSWORD` to the `openproxy-secrets` Secret.
3. Remove the `data` volume, its `volumeMount`, and `pvc.yaml`.
4. Raise `replicas` and switch the strategy back to `RollingUpdate`.

Migrations are embedded and run at startup against either driver.

## Security context

The pod runs as UID 65532 with a read-only root filesystem and all capabilities
dropped. `fsGroup` hands it ownership of the data volume, which — along with the
`/tmp` emptyDir — is the only path the process writes to.

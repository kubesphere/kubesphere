#!/usr/bin/env bash
set -euo pipefail

SELECTED_VERSION="${1:-}"
TARGET_CLUSTERS="${2:-}"
GW_API_EXISTS="${3:-false}"

if [ -z "$SELECTED_VERSION" ] || [ -z "$TARGET_CLUSTERS" ]; then
  echo "Usage: $0 <version> <clusters> [gw_api_exists]"
  echo "  clusters: space-separated list of cluster names (e.g. 'host' or 'host member1')"
  echo "  gw_api_exists: 'true' or 'false' (default: false)"
  exit 1
fi

INSTALLPLAN_FILE="/tmp/servicemesh-installplan.yaml"

cat > "$INSTALLPLAN_FILE" <<YAML
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: servicemesh
spec:
  extension:
    name: servicemesh
    version: "${SELECTED_VERSION}"
  enabled: true
  upgradeStrategy: Manual
  clusterScheduling:
    placement:
      clusters:
YAML

for cluster in $TARGET_CLUSTERS; do
  echo "      - \"${cluster}\"" >> "$INSTALLPLAN_FILE"
done

if [ "$GW_API_EXISTS" = "true" ]; then
  cat >> "$INSTALLPLAN_FILE" <<YAML
  config: |
    backend:
      istio:
        pilot:
          env:
            PILOT_ENABLE_GATEWAY_API: "false"
YAML
fi

echo "=== Generated InstallPlan ==="
cat "$INSTALLPLAN_FILE"
echo ""

echo "=== Dry-run validation ==="
kubectl apply --dry-run=server -f "$INSTALLPLAN_FILE"

echo ""
echo "✓ InstallPlan validated. Apply with: kubectl apply -f $INSTALLPLAN_FILE"
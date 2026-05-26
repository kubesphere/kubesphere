#!/usr/bin/env bash
set -euo pipefail

SELECTED_VERSION="${1:-${SELECTED_VERSION:-}}"
TARGET_CLUSTERS="${2:-${TARGET_CLUSTERS:-}}"

if [ -z "$SELECTED_VERSION" ] || [ -z "$TARGET_CLUSTERS" ]; then
  echo "Usage: $0 <version> <clusters>"
  echo "  clusters: space-separated list of cluster names (e.g. 'host' or 'host member1')"
  echo "  or set SELECTED_VERSION and TARGET_CLUSTERS env vars"
  exit 1
fi

if ! command -v kubectl &>/dev/null; then
  echo "Error: kubectl not found in PATH"
  exit 1
fi

INSTALLPLAN_FILE="/tmp/gateway-installplan.yaml"

cat > "$INSTALLPLAN_FILE" <<YAML
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: gateway
spec:
  extension:
    name: gateway
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

echo "=== Generated InstallPlan ==="
cat "$INSTALLPLAN_FILE"
echo ""

echo "=== Dry-run validation ==="
if ! kubectl apply --dry-run=server -f "$INSTALLPLAN_FILE"; then
  echo "✗ Dry-run failed — check version '${SELECTED_VERSION}' and cluster list."
  exit 1
fi

echo ""
echo "✓ InstallPlan validated. Apply with: kubectl apply -f $INSTALLPLAN_FILE"

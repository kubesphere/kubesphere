#!/usr/bin/env bash
set -euo pipefail

SELECTED_VERSION="${1:-}"

if [ -z "$SELECTED_VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "  version: KubeEye extension version (e.g. 1.0.1)"
  echo ""
  echo "Detect versions with:"
  echo "  kubectl get extensionversions -l kubesphere.io/extension-ref=kubeeye \\"
  echo "    -o jsonpath='{range .items[*]}{.spec.version}{\"\\n\"}{end}' | sort -V"
  exit 1
fi

INSTALLPLAN_FILE="/tmp/kubeeye-installplan.yaml"

cat > "$INSTALLPLAN_FILE" <<YAML
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: kubeeye
spec:
  extension:
    name: kubeeye
    version: "${SELECTED_VERSION}"
  enabled: true
  upgradeStrategy: Manual
YAML

echo "=== Generated InstallPlan ==="
cat "$INSTALLPLAN_FILE"
echo ""

echo "=== Dry-run validation ==="
kubectl apply --dry-run=server -f "$INSTALLPLAN_FILE"

echo ""
echo "✓ InstallPlan validated. Apply with: kubectl apply -f $INSTALLPLAN_FILE"

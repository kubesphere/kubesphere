#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-extension-gateway-api}"
TIMEOUT="${TIMEOUT:-300}"
ELAPSED=0

while [ $ELAPSED -lt $TIMEOUT ]; do
  if kubectl get installplans.kubesphere.io gateway-api &>/dev/null; then
    echo "  Waiting for InstallPlan to be deleted... (${ELAPSED}s)"
    sleep 10
    ELAPSED=$((ELAPSED + 10))
    continue
  fi

  REMAINING=$(kubectl get pods -n "$NAMESPACE" 2>/dev/null | grep -v -E 'uninstaller|Completed|NAME' | wc -l) || true

  if [ "$REMAINING" -eq 0 ]; then
    echo ""
    echo "✓ Uninstallation complete — all Gateway API components removed from $NAMESPACE."
    echo "  InstallPlan: deleted"
    echo "  Remaining pods: none"
    exit 0
  fi

  echo "  Waiting for components to be removed... (${ELAPSED}s, $REMAINING pods remaining)"
  sleep 10
  ELAPSED=$((ELAPSED + 10))
done

echo "⚠ Timeout reached. Remaining pods in $NAMESPACE:"
kubectl get pods -n "$NAMESPACE" 2>/dev/null | grep -v -E 'uninstaller|Completed'
exit 1

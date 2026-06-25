#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="extension-kubeeye"
TIMEOUT=300
ELAPSED=0

while [ $ELAPSED -lt $TIMEOUT ]; do
  if kubectl get installplans.kubesphere.io kubeeye --ignore-not-found &>/dev/null; then
    echo "⚠ InstallPlan still exists — uninstall may not have started."
    exit 1
  fi

  REMAINING=$(kubectl get pods -n "$NAMESPACE" 2>/dev/null | grep -v NAME | wc -l)

  if [ "$REMAINING" -eq 0 ]; then
    echo ""
    echo "✓ Uninstallation complete — all KubeEye components removed from $NAMESPACE."
    echo "  InstallPlan: deleted"
    echo "  Remaining pods: none"
    exit 0
  fi

  echo "  Waiting for components to be removed... (${ELAPSED}s, $REMAINING pods remaining)"
  sleep 10
  ELAPSED=$((ELAPSED + 10))
done

echo "⚠ Timeout reached. Remaining pods in $NAMESPACE:"
kubectl get pods -n "$NAMESPACE" 2>/dev/null | grep -v NAME
exit 1

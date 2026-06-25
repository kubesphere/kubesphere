#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="extension-kubeeye"
MODE="${1:-quick}"  # quick or poll

IP_NAME=$(kubectl get installplans.kubesphere.io kubeeye -o jsonpath='{.metadata.name}' --ignore-not-found 2>/dev/null || true)
if [ -z "$IP_NAME" ]; then
  echo "KubeEye InstallPlan not found. Is KubeEye installed?"
  exit 1
fi

if [ "$MODE" = "quick" ]; then
  echo "=== Extension State ==="
  echo "State: $(kubectl get installplans.kubesphere.io kubeeye -o jsonpath='{.status.state}' 2>/dev/null || true)"
  echo "Reason: $(kubectl get installplans.kubesphere.io kubeeye -o jsonpath='{.status.conditions[-1].reason}' 2>/dev/null || true)"
  echo "Message: $(kubectl get installplans.kubesphere.io kubeeye -o jsonpath='{.status.conditions[-1].message}' 2>/dev/null || true)"

  echo ""
  echo "=== Pod Status ==="
  kubectl get po -n "$NAMESPACE" -o wide 2>/dev/null || echo "No pods found in $NAMESPACE"
  exit 0
fi

# Polling mode
POLL_INTERVAL=10
TIMEOUT=300
ELAPSED=0

while [ $ELAPSED -lt $TIMEOUT ]; do
  EXT_STATE=$(kubectl get installplans.kubesphere.io kubeeye -o jsonpath='{.status.state}' 2>/dev/null)

  if [ -z "$EXT_STATE" ]; then
    echo "[$ELAPSED s] InstallPlan not found yet..."
  else
    echo "[$ELAPSED s] Extension: $EXT_STATE"

    if [ "$EXT_STATE" = "Installed" ]; then
      echo ""
      echo "✓ KubeEye installed successfully!"
      kubectl get po -n "$NAMESPACE" 2>/dev/null
      exit 0
    fi

    if [ "$EXT_STATE" = "Failed" ]; then
      echo "✗ Installation failed! Check details:"
      kubectl get installplans.kubesphere.io kubeeye -o yaml
      exit 1
    fi
  fi

  sleep $POLL_INTERVAL
  ELAPSED=$((ELAPSED + POLL_INTERVAL))
done

echo "⚠ Timeout reached ($TIMEOUT s). Current status:"
kubectl get installplans.kubesphere.io kubeeye -o yaml
exit 1

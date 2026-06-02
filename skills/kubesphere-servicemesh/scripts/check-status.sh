#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-quick}"  # quick or poll

if [ "$MODE" = "quick" ]; then
  echo "=== Extension State ==="
  echo "State: $(kubectl get installplans.kubesphere.io servicemesh -o jsonpath='{.status.state}')"
  echo "Reason: $(kubectl get installplans.kubesphere.io servicemesh -o jsonpath='{.status.conditions[-1].reason}')"
  echo "Message: $(kubectl get installplans.kubesphere.io servicemesh -o jsonpath='{.status.conditions[-1].message}')"

  echo ""
  echo "=== Per-Cluster Agent States ==="
  kubectl get installplans.kubesphere.io servicemesh -o go-template='
{{- range $cluster, $status := .status.clusterSchedulingStatuses}}
{{$cluster}}{{"\t"}}{{$status.state}}{{"\t"}}{{(index $status.conditions 0).reason}}{{"\n"}}
{{- end}}' | column -t
  exit 0
fi

# Polling mode
POLL_INTERVAL=10
TIMEOUT=300
ELAPSED=0

while [ $ELAPSED -lt $TIMEOUT ]; do
  EXT_STATE=$(kubectl get installplans.kubesphere.io servicemesh -o jsonpath='{.status.state}' 2>/dev/null)

  if [ -z "$EXT_STATE" ]; then
    echo "[$ELAPSED s] InstallPlan not found yet..."
  else
    AGENT_STATES=$(kubectl get installplans.kubesphere.io servicemesh \
      -o jsonpath='{.status.clusterSchedulingStatuses.*.state}')

    echo "[$ELAPSED s] Extension: $EXT_STATE | Agents: $AGENT_STATES"

    ALL_STATES="$EXT_STATE $AGENT_STATES"

    ALL_INSTALLED=true
    for state in $ALL_STATES; do
      if [ "$state" != "Installed" ]; then
        ALL_INSTALLED=false
        break
      fi
    done

    if [ "$ALL_INSTALLED" = true ]; then
      echo ""
      echo "✓ All ServiceMesh components installed successfully!"
      kubectl get installplans.kubesphere.io servicemesh -o go-template='
{{- range $cluster, $status := .status.clusterSchedulingStatuses}}
{{$cluster}}{{"\t"}}{{$status.state}}{{"\n"}}
{{- end}}' | column -t
      exit 0
    fi

    for state in $ALL_STATES; do
      if [ "$state" = "Failed" ]; then
        echo "✗ Installation failed! Check details:"
        kubectl get installplans.kubesphere.io servicemesh -o jsonpath='{.status}' | python3 -m json.tool
        exit 1
      fi
    done
  fi

  sleep $POLL_INTERVAL
  ELAPSED=$((ELAPSED + POLL_INTERVAL))
done

echo "⚠ Timeout reached ($TIMEOUT s). Current status:"
kubectl get installplans.kubesphere.io servicemesh -o jsonpath='{.status}' | python3 -m json.tool
exit 1
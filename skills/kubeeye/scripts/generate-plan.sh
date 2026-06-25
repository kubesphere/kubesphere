#!/usr/bin/env bash
set -euo pipefail

RULES=$(kubectl get inspectrule -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || true)

if [ -z "$RULES" ]; then
  echo "No InspectRules found. Run: kubectl apply -f rules/"
  exit 1
fi

{
  cat <<YAML
apiVersion: kubeeye.kubesphere.io/v1alpha2
kind: InspectPlan
metadata:
  name: inspectplan
spec:
  schedule: "* */12 * * ?"
  maxTasks: 10
  suspend: false
  timeout: 30m
  ruleNames:
YAML
  for rule in $RULES; do
    echo "  - name: $rule"
  done
} | kubectl apply -f -

echo "=== InspectPlan applied ==="
kubectl get inspectplan -o custom-columns=NAME:.metadata.name,RULES:.spec.ruleNames[*].name,SCHEDULE:.spec.schedule

#!/bin/bash

# generate-config.sh
# Generates whizard-telemetry config section based on installed extensions
# This script implements the same logic as generateConfig in installplan_controller.go

set -e

NAMESPACE="kubesphere-system"
LOGGING_NS="kubesphere-logging-system"

DEFAULT_OS_ENDPOINT="https://opensearch-cluster-data.kubesphere-logging-system.svc:9200"
DEFAULT_TIMESTRING="%Y.%m.%d"

# Function to check if extension is installed and enabled
check_extension() {
    local ext=$1
    local status=$(kubectl get installplan "$ext" -n "$NAMESPACE" -o jsonpath='{.spec.enabled}' 2>/dev/null || echo "")
    if [ "$status" = "true" ]; then
        return 0
    else
        return 1
    fi
}

# Get vector sink config from secret
get_vector_sink_config() {
    local secret_data=$(kubectl get secret vector-sinks -n "$LOGGING_NS" -o jsonpath='{.data.opensearch}' 2>/dev/null || echo "")
    if [ -n "$secret_data" ]; then
        echo "$secret_data" | base64 -d 2>/dev/null || echo ""
    fi
}

# Extract values from vector sink config (YAML format)
get_vector_endpoint() {
    local config="$1"
    # Extract the first endpoint from the YAML array
    echo "$config" | grep -A1 "endpoints:" | tail -1 | sed 's/^[[:space:]]*-[[:space:]]*//' | tr -d '\n'
}

get_vector_user() {
    local config="$1"
    echo "$config" | grep -A2 "auth:" | grep "user:" | sed 's/^[[:space:]]*user:[[:space:]]*//' | tr -d '\n'
}

get_vector_password() {
    local config="$1"
    echo "$config" | grep -A4 "auth:" | grep "password:" | sed 's/^[[:space:]]*password:[[:space:]]*//' | tr -d '\n'
}

# Get OS config values (from vector or defaults)
get_os_endpoint() {
    local vector_config="$1"
    local endpoint=$(get_vector_endpoint "$vector_config")
    if [ -z "$endpoint" ]; then
        echo "$DEFAULT_OS_ENDPOINT"
    else
        echo "$endpoint"
    fi
}

get_os_user() {
    local vector_config="$1"
    local user=$(get_vector_user "$vector_config")
    if [ -z "$user" ]; then
        echo "admin"
    else
        echo "$user"
    fi
}

get_os_password() {
    local vector_config="$1"
    local password=$(get_vector_password "$vector_config")
    if [ -z "$password" ]; then
        echo "admin"
    else
        echo "$password"
    fi
}

echo "Checking installed extensions..."

# Get vector config first
VECTOR_CONFIG=$(get_vector_sink_config)
OS_ENDPOINT=$(get_os_endpoint "$VECTOR_CONFIG")
OS_USER=$(get_os_user "$VECTOR_CONFIG")
OS_PASSWORD=$(get_os_password "$VECTOR_CONFIG")

if [ -n "$VECTOR_CONFIG" ]; then
    echo "  ✓ vector is installed (OpenSearch config from vector-sinks secret)"
else
    echo "  ✗ vector is NOT installed (using default OpenSearch config)"
fi

# Build config using array
declare -a CONFIG_LINES=()

# generateMonitoringConfig - checks whizard-monitoring and whizard-monitoring-pro
if check_extension "whizard-monitoring"; then
    echo "  ✓ whizard-monitoring is installed"
    CONFIG_LINES+=("monitoring:")
    CONFIG_LINES+=("  enable: true")
    if check_extension "whizard-monitoring-pro"; then
        echo "  ✓ whizard-monitoring-pro is installed"
        CONFIG_LINES+=("observability:")
        CONFIG_LINES+=("  enable: true")
    fi
fi

# generateLogsConfig - whizard-logging (uses generateMultiClusterConfig)
if check_extension "whizard-logging"; then
    echo "  ✓ whizard-logging is installed"
    CONFIG_LINES+=("logging:")
    CONFIG_LINES+=("  enable: true")
    CONFIG_LINES+=("  servers:")
    CONFIG_LINES+=("    - elasticsearch:")
    CONFIG_LINES+=("        endpoints:")
    CONFIG_LINES+=("          - ${OS_ENDPOINT}")
    CONFIG_LINES+=("        version: opensearchv2")
    CONFIG_LINES+=("        indexPrefix: \"{{ .cluster }}-logs\"")
    CONFIG_LINES+=("        timestring: \"${DEFAULT_TIMESTRING}\"")
    CONFIG_LINES+=("        basicAuth: true")
    CONFIG_LINES+=("        username: ${OS_USER}")
    CONFIG_LINES+=("        password: ${OS_PASSWORD}")
fi

# generateEventsConfig - whizard-events (uses generateMultiClusterConfig)
if check_extension "whizard-events"; then
    echo "  ✓ whizard-events is installed"
    CONFIG_LINES+=("events:")
    CONFIG_LINES+=("  enable: true")
    CONFIG_LINES+=("  servers:")
    CONFIG_LINES+=("    - elasticsearch:")
    CONFIG_LINES+=("        endpoints:")
    CONFIG_LINES+=("          - ${OS_ENDPOINT}")
    CONFIG_LINES+=("        version: opensearchv2")
    CONFIG_LINES+=("        indexPrefix: \"{{ .cluster }}-events\"")
    CONFIG_LINES+=("        timestring: \"${DEFAULT_TIMESTRING}\"")
    CONFIG_LINES+=("        basicAuth: true")
    CONFIG_LINES+=("        username: ${OS_USER}")
    CONFIG_LINES+=("        password: ${OS_PASSWORD}")
fi

# generateAuditingConfig - whizard-auditing (uses generateMultiClusterConfig)
if check_extension "whizard-auditing"; then
    echo "  ✓ whizard-auditing is installed"
    CONFIG_LINES+=("auditing:")
    CONFIG_LINES+=("  enable: true")
    CONFIG_LINES+=("  servers:")
    CONFIG_LINES+=("    - elasticsearch:")
    CONFIG_LINES+=("        endpoints:")
    CONFIG_LINES+=("          - ${OS_ENDPOINT}")
    CONFIG_LINES+=("        version: opensearchv2")
    CONFIG_LINES+=("        indexPrefix: \"{{ .cluster }}-auditing\"")
    CONFIG_LINES+=("        timestring: \"${DEFAULT_TIMESTRING}\"")
    CONFIG_LINES+=("        basicAuth: true")
    CONFIG_LINES+=("        username: ${OS_USER}")
    CONFIG_LINES+=("        password: ${OS_PASSWORD}")
fi

# generateNotificationConfig - whizard-notification
if check_extension "whizard-notification"; then
    echo "  ✓ whizard-notification is installed"
    CONFIG_LINES+=("notification:")
    CONFIG_LINES+=("  endpoint: http://notification-manager-svc.kubesphere-monitoring-system.svc:19093")
    CONFIG_LINES+=("  history:")
    CONFIG_LINES+=("    enable: true")
    CONFIG_LINES+=("    server:")
    CONFIG_LINES+=("      elasticsearch:")
    CONFIG_LINES+=("        endpoints:")
    CONFIG_LINES+=("          - ${OS_ENDPOINT}")
    CONFIG_LINES+=("        version: opensearchv2")
    CONFIG_LINES+=("        indexPrefix: \"{{ .cluster }}-notification-history\"")
    CONFIG_LINES+=("        timestring: \"${DEFAULT_TIMESTRING}\"")
    CONFIG_LINES+=("        basicAuth: true")
    CONFIG_LINES+=("        username: ${OS_USER}")
    CONFIG_LINES+=("        password: ${OS_PASSWORD}")
fi

# generateTracingConfig - wiztelemetry-tracing
# Note: tracing uses its own InstallPlan config, NOT from vector secret
# Config structure: global.storage.opensearch.index.prefix and global.storage.opensearch.index.timestring
if check_extension "wiztelemetry-tracing"; then
    echo "  ✓ wiztelemetry-tracing is installed"
    # Default values
    TRACING_OS_ENDPOINT="https://opensearch-cluster-data.kubesphere-logging-system.svc:9200"
    TRACING_OS_USER="admin"
    TRACING_OS_PASSWORD="admin"
    TRACING_SPAN_PREFIX="wiz-tracing-span"
    TRACING_SPAN_TIMESTRING="$DEFAULT_TIMESTRING"
    
    # Get tracing config from its InstallPlan
    tracing_config=$(kubectl get installplan wiztelemetry-tracing -n "$NAMESPACE" -o jsonpath='{.spec.config}' 2>/dev/null || echo "")
    if [ -n "$tracing_config" ]; then
        # Extract endpoint
        custom_endpoint=$(echo "$tracing_config" | grep -A1 "endpoints:" | tail -1 | sed 's/^[[:space:]]*-[[:space:]]*//' | tr -d '\n')
        [ -n "$custom_endpoint" ] && TRACING_OS_ENDPOINT="$custom_endpoint"
        
        # Extract auth
        custom_user=$(echo "$tracing_config" | grep -A2 "user:" | head -1 | sed 's/^[[:space:]]*user:[[:space:]]*//' | tr -d '\n')
        custom_password=$(echo "$tracing_config" | grep -A1 "password:" | head -1 | sed 's/^[[:space:]]*password:[[:space:]]*//' | tr -d '\n')
        [ -n "$custom_user" ] && TRACING_OS_USER="$custom_user"
        [ -n "$custom_password" ] && TRACING_OS_PASSWORD="$custom_password"
        
        # Extract index prefix from global.storage.opensearch.index
        custom_span_prefix=$(echo "$tracing_config" | grep -A3 "index:" | grep "prefix:" | sed 's/^[[:space:]]*prefix:[[:space:]]*//' | tr -d '\n' | tr -d '"')
        [ -n "$custom_span_prefix" ] && TRACING_SPAN_PREFIX="$custom_span_prefix"

        custom_span_timestring=$(echo "$tracing_config" | grep -A3 "index:" | grep "timestring:" | sed 's/^[[:space:]]*timestring:[[:space:]]*//' | tr -d '\n' | tr -d '"')
        [ -n "$custom_span_timestring" ] && TRACING_SPAN_TIMESTRING="$custom_span_timestring"
    fi
    CONFIG_LINES+=("tracing:")
    CONFIG_LINES+=("  enable: true")
    CONFIG_LINES+=("  server:")
    CONFIG_LINES+=("    elasticsearch:")
    CONFIG_LINES+=("      endpoints:")
    CONFIG_LINES+=("        - ${TRACING_OS_ENDPOINT}")
    CONFIG_LINES+=("      basicAuth: true")
    CONFIG_LINES+=("      username: ${TRACING_OS_USER}")
    CONFIG_LINES+=("      password: ${TRACING_OS_PASSWORD}")
    CONFIG_LINES+=("      version: opensearchv2")
    CONFIG_LINES+=("  spanIndex:")
    CONFIG_LINES+=("    prefix: ${TRACING_SPAN_PREFIX}")
    CONFIG_LINES+=("    timeString: \"${TRACING_SPAN_TIMESTRING}\"")
fi

# generateEbpfConfig - wiztelemetry-bpfconductor (uses generateMultiClusterConfig)
if check_extension "wiztelemetry-bpfconductor"; then
    echo "  ✓ wiztelemetry-bpfconductor is installed"
    CONFIG_LINES+=("ebpf:")
    CONFIG_LINES+=("  httpTraffic:")
    CONFIG_LINES+=("    enable: true")
    CONFIG_LINES+=("    servers:")
    CONFIG_LINES+=("      - elasticsearch:")
    CONFIG_LINES+=("          endpoints:")
    CONFIG_LINES+=("            - ${OS_ENDPOINT}")
    CONFIG_LINES+=("          version: opensearchv2")
    CONFIG_LINES+=("          indexPrefix: \"{{ .cluster }}-traffic\"")
    CONFIG_LINES+=("          timestring: \"${DEFAULT_TIMESTRING}\"")
    CONFIG_LINES+=("          basicAuth: true")
    CONFIG_LINES+=("          username: ${OS_USER}")
    CONFIG_LINES+=("          password: ${OS_PASSWORD}")
fi

# generateEventsAlertingConfig - whizard-telemetry-ruler
# Check: global.alertingPersistence.enabled must be true
if check_extension "whizard-telemetry-ruler"; then
    echo "  ✓ whizard-telemetry-ruler is installed"
    # Get config and check if alertingPersistence is enabled
    ruler_config=$(kubectl get installplan whizard-telemetry-ruler -n "$NAMESPACE" -o jsonpath='{.spec.config}' 2>/dev/null || echo "")
    EA_ENABLED="false"
    EA_OS_ENABLED="true"
    EA_OS_PREFIX="{{ .cluster }}-events-alerting"
    EA_OS_TIMESTRING="$DEFAULT_TIMESTRING"
    
    if [ -n "$ruler_config" ]; then
        # Check global.alertingPersistence.enabled
        global_enabled=$(echo "$ruler_config" | grep -A2 "alertingPersistence:" | grep "enabled:" | sed 's/^[[:space:]]*enabled:[[:space:]]*//' | tr -d '\n')
        [ "$global_enabled" = "true" ] && EA_ENABLED="true"
        
        # Check alertingPersistence.sinks.opensearch.enabled
        os_enabled=$(echo "$ruler_config" | grep -A10 "sinks:" | grep -A2 "opensearch:" | grep "enabled:" | sed 's/^[[:space:]]*enabled:[[:space:]]*//' | tr -d '\n')
        [ -n "$os_enabled" ] && EA_OS_ENABLED="$os_enabled"
        
        # Check index prefix
        os_prefix=$(echo "$ruler_config" | grep -A10 "sinks:" | grep -A5 "opensearch:" | grep "prefix:" | sed 's/^[[:space:]]*prefix:[[:space:]]*//' | tr -d '\n' | tr -d '"')
        [ -n "$os_prefix" ] && EA_OS_PREFIX="$os_prefix"
        
        os_timestring=$(echo "$ruler_config" | grep -A10 "sinks:" | grep -A5 "opensearch:" | grep "timestring:" | sed 's/^[[:space:]]*timestring:[[:space:]]*//' | tr -d '\n' | tr -d '"')
        [ -n "$os_timestring" ] && EA_OS_TIMESTRING="$os_timestring"
    else
        # Default: enabled=false
        EA_ENABLED="false"
    fi
    
    if [ "$EA_ENABLED" = "true" ] && [ "$EA_OS_ENABLED" = "true" ]; then
        CONFIG_LINES+=("eventsAlerting:")
        CONFIG_LINES+=("  enable: true")
        CONFIG_LINES+=("  servers:")
        CONFIG_LINES+=("    - elasticsearch:")
        CONFIG_LINES+=("        endpoints:")
        CONFIG_LINES+=("          - ${OS_ENDPOINT}")
        CONFIG_LINES+=("        version: opensearchv2")
        CONFIG_LINES+=("        basicAuth: true")
        CONFIG_LINES+=("        username: ${OS_USER}")
        CONFIG_LINES+=("        password: ${OS_PASSWORD}")
        CONFIG_LINES+=("        indexPrefix: \"${EA_OS_PREFIX}\"")
        CONFIG_LINES+=("        timestring: \"${EA_OS_TIMESTRING}\"")
    fi
fi

# generateMetricsAlertingConfig - whizard-alerting
# Check: alertingPersistence.enabled must be true
if check_extension "whizard-alerting"; then
    echo "  ✓ whizard-alerting is installed"
    # Get config and check if alertingPersistence is enabled
    alert_config=$(kubectl get installplan whizard-alerting -n "$NAMESPACE" -o jsonpath='{.spec.config}' 2>/dev/null || echo "")
    MA_ENABLED="false"
    MA_OS_ENABLED="true"
    MA_OS_PREFIX="{{ .cluster }}-metrics-alerting"
    MA_OS_TIMESTRING="$DEFAULT_TIMESTRING"
    
    if [ -n "$alert_config" ]; then
        # Check alertingPersistence.enabled (can be "alerting-persistence" or "alertingPersistence")
        persistence_enabled=$(echo "$alert_config" | grep -E "(alerting-persistence|alertingPersistence):" -A1 | grep "enabled:" | sed 's/^[[:space:]]*enabled:[[:space:]]*//' | tr -d '\n')
        [ "$persistence_enabled" = "true" ] && MA_ENABLED="true"
        
        # Check alertingPersistence.sinks.opensearch.enabled
        os_enabled=$(echo "$alert_config" | grep "opensearch:" -A2 | grep "enabled:" | sed 's/^[[:space:]]*enabled:[[:space:]]*//' | tr -d '\n')
        [ -n "$os_enabled" ] && MA_OS_ENABLED="$os_enabled"
        
        # Check index prefix
        os_prefix=$(echo "$alert_config" | grep -A10 "sinks:" | grep -A5 "opensearch:" | grep "prefix:" | sed 's/^[[:space:]]*prefix:[[:space:]]*//' | tr -d '\n' | tr -d '"')
        [ -n "$os_prefix" ] && MA_OS_PREFIX="$os_prefix"
        
        os_timestring=$(echo "$alert_config" | grep -A10 "sinks:" | grep -A5 "opensearch:" | grep "timestring:" | sed 's/^[[:space:]]*timestring:[[:space:]]*//' | tr -d '\n' | tr -d '"')
        [ -n "$os_timestring" ] && MA_OS_TIMESTRING="$os_timestring"
    else
        # Default: enabled=false
        MA_ENABLED="false"
    fi
    
    if [ "$MA_ENABLED" = "true" ] && [ "$MA_OS_ENABLED" = "true" ]; then
        CONFIG_LINES+=("metricsAlerting:")
        CONFIG_LINES+=("  enable: true")
        CONFIG_LINES+=("  server:")
        CONFIG_LINES+=("    elasticsearch:")
        CONFIG_LINES+=("      endpoints:")
        CONFIG_LINES+=("        - ${OS_ENDPOINT}")
        CONFIG_LINES+=("      basicAuth: true")
        CONFIG_LINES+=("      username: ${OS_USER}")
        CONFIG_LINES+=("      password: ${OS_PASSWORD}")
        CONFIG_LINES+=("      indexPrefix: \"${MA_OS_PREFIX}\"")
        CONFIG_LINES+=("      timestring: \"${MA_OS_TIMESTRING}\"")
    fi
fi

if [ ${#CONFIG_LINES[@]} -eq 0 ]; then
    echo "No observability extensions installed"
fi

# Output the config section - formatted for direct use in InstallPlan config: | block
echo ""
echo "---CONFIG_START---"
echo "whizard-telemetry:"
echo "  config: "
for line in "${CONFIG_LINES[@]}"; do
    echo "    $line"
done
echo "---CONFIG_END---"

echo ""
echo "OpenSearch config used:"
echo "  Endpoint: ${OS_ENDPOINT}"
echo "  User: ${OS_USER}"
echo "  Password: ${OS_PASSWORD}"

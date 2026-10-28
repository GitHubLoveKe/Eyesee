#!/bin/bash
# Integration test for Zone Coordinator

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0

test_api() {
    local desc="$1"
    local method="$2"
    local path="$3"
    local data="$4"
    local expect="$5"

    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        result=$(curl -s -X POST "$BASE_URL$path" -H "Content-Type: application/json" -d "$data")
    elif [ "$method" = "POST" ]; then
        result=$(curl -s -X POST "$BASE_URL$path")
    else
        result=$(curl -s "$BASE_URL$path")
    fi

    if echo "$result" | grep -q "$expect"; then
        echo "✓ $desc"
        PASS=$((PASS+1))
    else
        echo "✗ $desc"
        echo "  Expected: $expect"
        echo "  Got: $result"
        FAIL=$((FAIL+1))
    fi
}

echo "=== Zone Coordinator Integration Test ==="
echo ""

# 1. Health check
test_api "Health check" GET "/health" "" "OK"

# 2. Register vmagent worker
test_api "Register vmagent worker" POST "/api/v1/workers/register" \
    '{"worker_id":"vmagent-1","worker_type":"vmagent","address":"http://192.168.1.10:8429","labels":{"dc":"zone-a"}}' \
    '"code":0'

# 3. Register vmalert worker
test_api "Register vmalert worker" POST "/api/v1/workers/register" \
    '{"worker_id":"vmalert-1","worker_type":"vmalert","address":"http://192.168.1.11:8880","labels":{"dc":"zone-a"}}' \
    '"code":0'

# 4. Push resources
test_api "Push resources" POST "/api/v1/push/resources" \
    '[{"id":"mysql-1","type":"mysql","address":"10.0.1.5","port":3306,"labels":{"env":"prod"}}]' \
    '"code":0'

# 5. Push collection tasks
test_api "Push collection tasks" POST "/api/v1/push/tasks" \
    '[{"task_id":"task-mysql-1","instance_id":"mysql-1","name":"MySQL Metrics","job_name":"mysql-exporter","target":"10.0.1.5:9104","metrics_path":"/metrics","scrape_interval":"15s","enabled":true}]' \
    '"code":0'

# 6. Push alert rules
test_api "Push alert rules" POST "/api/v1/push/rules" \
    '[{"rule_id":"rule-mysql-1","task_id":"task-mysql-1","alert_name":"MySQLDown","expr":"mysql_up == 0","duration":"5m","severity":"critical"}]' \
    '"code":0'

# 7. Check status
test_api "Check status" GET "/api/v1/status" "" '"total_workers":2'

# 8. Check assignments
test_api "Check assignments" GET "/api/v1/assignments" "" '"assignments"'

# 9. Get vmagent scrape config
test_api "Get vmagent scrape config" GET "/api/v1/workers/vmagent-1/scrape-config" "" "scrape_configs"

# 10. Get vmalert rules
test_api "Get vmalert rules" GET "/api/v1/workers/vmalert-1/rules" "" "groups"

# 11. Worker heartbeat
test_api "Worker heartbeat" POST "/api/v1/workers/vmagent-1/heartbeat" \
    '{"timestamp":"2026-10-28T11:00:00Z","status":"healthy","metrics":{"cpu":0.5}}' \
    '"code":0'

# 12. Worker assignments
test_api "Worker assignments" GET "/api/v1/workers/vmagent-1/assignments" "" '"task_ids"'

echo ""
echo "=== Test Summary ==="
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "All tests passed!"
    exit 0
else
    echo "Some tests failed!"
    exit 1
fi

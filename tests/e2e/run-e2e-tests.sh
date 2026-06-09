#!/bin/bash

# End-to-end test suite for Mini Kubernet
# Tests authentication flow, OTP, deployment, and service connectivity

set -e

ENVIRONMENT=${1:-dev}
NAMESPACE="kubernet-${ENVIRONMENT}"
TESTS_PASSED=0
TESTS_FAILED=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

# Test 1: Check namespace exists
print_test "Checking if namespace $NAMESPACE exists"
if kubectl get namespace "$NAMESPACE" &>/dev/null; then
    print_pass "Namespace exists"
else
    print_fail "Namespace does not exist"
fi

# Test 2: Check all deployments are running
print_test "Checking deployments are running"
READY_REPLICAS=$(kubectl get deployments -n "$NAMESPACE" -o jsonpath='{.items[*].status.readyReplicas}' | awk '{sum=0; for(i=1;i<=NF;i++) sum+=$i} END {print sum}')
TOTAL_REPLICAS=$(kubectl get deployments -n "$NAMESPACE" -o jsonpath='{.items[*].spec.replicas}' | awk '{sum=0; for(i=1;i<=NF;i++) sum+=$i} END {print sum}')

if [ "$READY_REPLICAS" -eq "$TOTAL_REPLICAS" ]; then
    print_pass "All $TOTAL_REPLICAS replicas are ready"
else
    print_fail "Only $READY_REPLICAS/$TOTAL_REPLICAS replicas are ready"
fi

# Test 3: Check services are accessible
print_test "Checking services are accessible"
for service in auth-service api-gateway frontend; do
    if kubectl get service "$service" -n "$NAMESPACE" &>/dev/null; then
        print_pass "Service $service exists"
    else
        print_fail "Service $service not found"
    fi
done

# Test 4: Check pod logs for errors
print_test "Checking pod logs for errors"
ERROR_COUNT=$(kubectl logs -n "$NAMESPACE" --all-containers=true --timestamps=false --prefix=true 2>/dev/null | grep -i "error" | wc -l || echo 0)
if [ "$ERROR_COUNT" -eq 0 ]; then
    print_pass "No errors found in logs"
else
    print_fail "Found $ERROR_COUNT errors in logs"
fi

# Test 5: Check database connectivity
print_test "Checking database connectivity"
# This would require a test pod with postgres client
# For now, we'll check if RDS endpoint is resolvable
if kubectl run -it --rm --image=busybox --restart=Never -n "$NAMESPACE" test-db --command -- nslookup postgres 2>/dev/null | grep -q "postgres"; then
    print_pass "Database is resolvable"
else
    print_fail "Database is not resolvable"
fi

# Test 6: Check Redis connectivity
print_test "Checking Redis connectivity"
if kubectl run -it --rm --image=busybox --restart=Never -n "$NAMESPACE" test-redis --command -- nslookup redis 2>/dev/null | grep -q "redis"; then
    print_pass "Redis is resolvable"
else
    print_fail "Redis is not resolvable"
fi

# Test 7: Run API health check
print_test "Running API health check"
HEALTH_CHECK=$(kubectl exec -n "$NAMESPACE" -it $(kubectl get pod -n "$NAMESPACE" -l app=api-gateway -o jsonpath='{.items[0].metadata.name}') -- curl -s http://localhost:8080/health 2>/dev/null || echo "FAIL")

if [ "$HEALTH_CHECK" != "FAIL" ]; then
    print_pass "API health check passed"
else
    print_fail "API health check failed"
fi

# Test 8: Check persistent volumes
print_test "Checking persistent volumes"
PV_COUNT=$(kubectl get pvc -n "$NAMESPACE" -o jsonpath='{.items[*].status.phase}' | grep -o "Bound" | wc -l || echo 0)
if [ "$PV_COUNT" -gt 0 ]; then
    print_pass "Found $PV_COUNT bound persistent volumes"
else
    print_fail "No bound persistent volumes found"
fi

# Test 9: Check metrics are being scraped
print_test "Checking metrics collection"
METRICS=$(kubectl port-forward -n monitoring svc/prometheus 9090:9090 &>/dev/null & sleep 2; curl -s http://localhost:9090/api/v1/targets | grep -o '"health":"up"' | wc -l; pkill -f "port-forward" || true)

if [ "$METRICS" -gt 0 ]; then
    print_pass "Metrics collection active ($METRICS targets up)"
else
    print_fail "Metrics collection not working"
fi

# Print summary
echo ""
echo "========================================"
echo "E2E Test Summary"
echo "========================================"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi

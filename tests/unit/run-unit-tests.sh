#!/bin/bash

# Unit test runner for backend services
set -e

SERVICES=("auth-service" "api-gateway" "build-service" "deploy-service")

echo "Running unit tests for Mini Kubernet backend services..."
echo "========================================"

TOTAL_TESTS=0
PASSED_TESTS=0

for service in "${SERVICES[@]}"; do
    SERVICE_DIR="backend/$service"
    
    if [ ! -d "$SERVICE_DIR" ]; then
        echo "✗ Service directory not found: $SERVICE_DIR"
        continue
    fi
    
    echo ""
    echo "Testing $service..."
    
    cd "$SERVICE_DIR"
    
    # Run Go tests
    TEST_OUTPUT=$(go test -v -race -coverprofile=coverage.out ./... 2>&1 || true)
    
    # Parse test results
    if echo "$TEST_OUTPUT" | grep -q "PASS"; then
        COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}')
        echo "✓ $service tests passed (coverage: $COVERAGE)"
        ((PASSED_TESTS++))
    else
        echo "✗ $service tests failed"
        echo "$TEST_OUTPUT"
    fi
    
    ((TOTAL_TESTS++))
    cd - > /dev/null
done

echo ""
echo "========================================"
echo "Test Summary: $PASSED_TESTS/$TOTAL_TESTS services passed"

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo "✓ All unit tests passed!"
    exit 0
else
    echo "✗ Some tests failed"
    exit 1
fi

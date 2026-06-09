# Incident Response & Disaster Recovery Guide

## Table of Contents
1. [Incident Response Procedures](#incident-response-procedures)
2. [Common Incidents & Resolution](#common-incidents--resolution)
3. [Disaster Recovery Procedures](#disaster-recovery-procedures)
4. [Post-Incident Review](#post-incident-review)

## Incident Response Procedures

### Incident Severity Levels

**CRITICAL (SEV-1)**
- Complete service outage
- Data loss or corruption
- Security breach
- Response time: 15 minutes
- Escalation: VP Engineering + CTO

**HIGH (SEV-2)**
- Partial service degradation
- Increased error rate > 10%
- Performance degradation > 50%
- Response time: 1 hour
- Escalation: Engineering Manager + On-call Engineer

**MEDIUM (SEV-3)**
- Minor functionality impaired
- Error rate 5-10%
- Performance degradation 10-50%
- Response time: 4 hours
- Escalation: On-call Engineer

**LOW (SEV-4)**
- Cosmetic issues or minor bugs
- Minimal user impact
- Response time: 24 hours
- Escalation: Normal ticketing process

### Incident Response Workflow

```
1. Detection & Alerting
   ↓
2. Triage & Classification
   ↓
3. Immediate Mitigation
   ↓
4. Root Cause Analysis
   ↓
5. Resolution & Deployment
   ↓
6. Verification & Monitoring
   ↓
7. Communication & Documentation
```

### Immediate Actions

1. **Page On-call Engineer** (via PagerDuty)
2. **Create Incident Ticket** (Jira)
3. **Notify Stakeholders** (Slack #incidents channel)
4. **Start War Room** (Conference bridge)
5. **Assess Impact** (Services down, user impact)

### Mitigation Strategies

**For Authentication Service Down**
```bash
# 1. Check pod status
kubectl get pods -n kubernet-prod -l app=auth-service

# 2. Check logs for errors
kubectl logs -n kubernet-prod -l app=auth-service --tail=100

# 3. Check resource limits
kubectl describe pod <pod-name> -n kubernet-prod

# 4. If pod is in crashloop, scale down and investigate
kubectl scale deployment auth-service --replicas=0 -n kubernet-prod

# 5. Once fixed, scale back up
kubectl scale deployment auth-service --replicas=5 -n kubernet-prod
```

**For API Gateway Down**
```bash
# 1. Check ingress status
kubectl get ingress -n kubernet-prod

# 2. Check service endpoints
kubectl get endpoints api-gateway -n kubernet-prod

# 3. Check pod logs
kubectl logs -n kubernet-prod -l app=api-gateway --tail=100

# 4. Perform rolling restart
kubectl rollout restart deployment/api-gateway -n kubernet-prod
```

**For Database Connection Issues**
```bash
# 1. Check RDS status
aws rds describe-db-instances --db-instance-identifier mini-kubernet-prod

# 2. Check connection pool
kubectl exec -n kubernet-prod <pod> -- psql -h postgres.kubernet-prod.svc.cluster.local -c "SELECT count(*) FROM pg_stat_activity;"

# 3. Scale down connections if needed
kubectl set env deployment/auth-service DB_POOL_SIZE=10 -n kubernet-prod

# 4. Perform database restart if necessary
aws rds reboot-db-instance --db-instance-identifier mini-kubernet-prod
```

**For High Memory Usage**
```bash
# 1. Identify pod with high memory
kubectl top pods -n kubernet-prod

# 2. Check if OOMKilled
kubectl describe pod <pod-name> -n kubernet-prod | grep -i oom

# 3. Increase memory limit
kubectl set resources deployment/<service> --limits=memory=2Gi -n kubernet-prod

# 4. Monitor memory usage
kubectl top pod <pod-name> -n kubernet-prod --containers
```

## Common Incidents & Resolution

### Pod CrashLoopBackOff

**Symptoms**: Pod repeatedly crashes and restarts

**Diagnosis**:
```bash
kubectl logs -n kubernet-prod <pod-name> --tail=50
kubectl describe pod <pod-name> -n kubernet-prod
```

**Common Causes & Solutions**:

1. **Missing Environment Variables**
   ```bash
   kubectl set env deployment/<name> KEY=value -n kubernet-prod
   ```

2. **Health Check Failing**
   ```bash
   # Temporarily disable health checks
   kubectl set probe deployment/<name> --liveness --initial-delay-seconds=60 -n kubernet-prod
   ```

3. **Out of Memory**
   ```bash
   kubectl set resources deployment/<name> --limits=memory=1Gi -n kubernet-prod
   ```

4. **Application Error**
   ```bash
   # Review application logs
   kubectl logs <pod-name> -n kubernet-prod -c <container-name>
   # Check for stack traces or error messages
   ```

### High Latency

**Symptoms**: Requests taking > 1 second, p95 latency > 500ms

**Diagnosis**:
```bash
# Check metrics in Prometheus
# Query: histogram_quantile(0.95, rate(http_request_duration_ms[5m]))

# Check database query performance
kubectl exec -n kubernet-prod <postgres-pod> -- \
  psql -c "SELECT query, calls, mean_time FROM pg_stat_statements ORDER BY mean_time DESC;"
```

**Solutions**:

1. **Database Optimization**
   ```bash
   # Analyze slow queries
   kubectl exec -n kubernet-prod <postgres-pod> -- \
     psql -c "ANALYZE; EXPLAIN ANALYZE SELECT ...;"
   
   # Create missing indexes
   kubectl exec -n kubernet-prod <postgres-pod> -- \
     psql -c "CREATE INDEX idx_name ON table(column);"
   ```

2. **Scale Services**
   ```bash
   kubectl scale deployment/<service> --replicas=10 -n kubernet-prod
   ```

3. **Add Caching**
   ```bash
   # Increase Redis replicas
   kubectl scale deployment redis --replicas=3 -n kubernet-prod
   ```

### High Error Rate

**Symptoms**: Error rate > 5%, increased 5xx responses

**Diagnosis**:
```bash
# Check application logs
kubectl logs -n kubernet-prod -l app=<service> --since=10m | grep -i error

# Check metrics
# Query: rate(http_requests_total{status=~"5.."}[5m])
```

**Solutions**:

1. **For Authentication Errors**
   ```bash
   # Check JWT secret
   kubectl get secret -n kubernet-prod auth-service-secrets -o yaml

   # Check database connectivity
   kubectl exec -n kubernet-prod <pod> -- \
     psql -h postgres.kubernet-prod.svc.cluster.local -c "SELECT 1;"
   ```

2. **For Database Errors**
   ```bash
   # Check connection pool
   kubectl logs -n kubernet-prod -l app=auth-service | grep -i "connection"

   # Increase pool size
   kubectl set env deployment/auth-service DB_POOL_SIZE=50 -n kubernet-prod
   ```

3. **For Timeout Errors**
   ```bash
   # Increase timeout values
   kubectl set env deployment/<service> \
     REQUEST_TIMEOUT=30s \
     DB_TIMEOUT=10s \
     -n kubernet-prod
   ```

### Disk Space Issues

**Symptoms**: Pod eviction, "No space left on device" errors

**Diagnosis**:
```bash
kubectl exec -n kubernet-prod <pod> -- df -h
kubectl exec -n kubernet-prod <pod> -- du -sh /*
```

**Solutions**:

```bash
# 1. Clean up logs
kubectl exec -n kubernet-prod <pod> -- rm -rf /var/log/*

# 2. Clean up cache
kubectl exec -n kubernet-prod <pod> -- rm -rf /app/cache/*

# 3. Increase PVC size
kubectl patch pvc <pvc-name> -n kubernet-prod -p '{"spec":{"resources":{"requests":{"storage":"50Gi"}}}}'

# 4. Enable log rotation
kubectl set env deployment/<service> LOG_MAX_SIZE=100M -n kubernet-prod
```

## Disaster Recovery Procedures

### Complete Cluster Failure

**RTO**: 1 hour | **RPO**: 5 minutes

**Recovery Steps**:

1. **Provision New Cluster**
   ```bash
   terraform apply -var-file="environments/prod/terraform.tfvars"
   aws eks update-kubeconfig --name mini-kubernet-prod
   ```

2. **Restore from Latest Snapshot**
   ```bash
   aws rds restore-db-instance-from-db-snapshot \
     --db-instance-identifier mini-kubernet-prod-restored \
     --db-snapshot-identifier mini-kubernet-prod-latest
   ```

3. **Deploy Applications**
   ```bash
   ./scripts/deployment/deploy.sh prod
   ```

4. **Restore Data**
   ```bash
   aws s3 cp s3://mini-kubernet-backups/prod/latest/ ./backups/ --recursive
   kubectl cp ./backups/vault-snapshot.db vault/vault-0:/tmp/
   kubectl exec -n vault vault-0 -- vault operator raft restore /tmp/vault-snapshot.db
   ```

5. **Verify Services**
   ```bash
   ./tests/e2e/run-e2e-tests.sh prod
   ```

### Database Failure

1. **Failover to Read Replica**
   ```bash
   aws rds promote-read-replica --db-instance-identifier mini-kubernet-prod-read-replica
   ```

2. **Update Application Connection String**
   ```bash
   kubectl set env deployment/auth-service \
     DB_HOST=mini-kubernet-prod-promoted.xxxxx.rds.amazonaws.com \
     -n kubernet-prod
   ```

3. **Verify Connectivity**
   ```bash
   kubectl exec -n kubernet-prod <pod> -- \
     psql -h <new-endpoint> -c "SELECT 1;"
   ```

### Data Corruption

1. **Restore from Point-in-Time Backup**
   ```bash
   # Find backup before corruption
   aws rds describe-db-snapshots \
     --db-instance-identifier mini-kubernet-prod \
     --query 'DBSnapshots[?SnapshotCreateTime>`2024-01-01`]'

   # Restore to specific point in time
   aws rds restore-db-instance-to-point-in-time \
     --source-db-instance-identifier mini-kubernet-prod \
     --target-db-instance-identifier mini-kubernet-prod-recovered \
     --restore-time 2024-01-15T14:30:00Z
   ```

2. **Swap to Recovered Instance**
   ```bash
   # Update DNS/connection string to point to recovered instance
   kubectl set env deployment/auth-service \
     DB_HOST=mini-kubernet-prod-recovered.xxxxx.rds.amazonaws.com \
     -n kubernet-prod
   ```

### Network Partition / Split Brain

1. **Identify Affected Services**
   ```bash
   kubectl get nodes
   kubectl get pods --all-namespaces -o wide
   ```

2. **Drain Unhealthy Nodes**
   ```bash
   kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data
   ```

3. **Replace Node**
   ```bash
   kubectl delete node <node-name>
   # AWS Auto Scaling will automatically provision replacement
   ```

## Post-Incident Review

### Template for Incident Report

```
# Incident Report: [INCIDENT-ID]

## Summary
- **Date**: YYYY-MM-DD
- **Duration**: HH:MM
- **Severity**: SEV-1/2/3/4
- **Status**: RESOLVED

## Timeline
| Time | Event |
|------|-------|
| HH:MM | Detection |
| HH:MM | Root cause identified |
| HH:MM | Mitigation applied |
| HH:MM | Verified resolved |

## Root Cause
[Describe what happened and why]

## Impact
- Users affected: [number]
- Services impacted: [list]
- Revenue impact: $[amount]

## Resolution
[Describe what was done to fix the issue]

## Action Items
- [ ] Implement permanent fix by [date]
- [ ] Add monitoring/alerting for early detection
- [ ] Update runbooks
- [ ] Training for team

## Prevention
[What will prevent this from happening again]

## Lessons Learned
[Key takeaways and improvements]
```

### Review Schedule
- **SEV-1**: 24 hours
- **SEV-2**: 48 hours
- **SEV-3**: 1 week
- **SEV-4**: As needed

### Continuous Improvement
1. Share findings with team
2. Update documentation and runbooks
3. Implement preventive measures
4. Track action items to completion
5. Monthly review of all incidents

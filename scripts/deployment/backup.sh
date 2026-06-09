#!/bin/bash

# Backup script for Mini Kubernet
# Backs up databases, persistent volumes, and configurations

set -e

ENVIRONMENT=${1:-prod}
BACKUP_DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_DIR="backups/${ENVIRONMENT}/${BACKUP_DATE}"

mkdir -p "$BACKUP_DIR"

print_info() {
    echo "[INFO] $1"
}

# Backup RDS database
print_info "Creating RDS snapshot..."
aws rds create-db-snapshot \
    --db-instance-identifier mini-kubernet-${ENVIRONMENT} \
    --db-snapshot-identifier mini-kubernet-${ENVIRONMENT}-${BACKUP_DATE}

# Backup persistent volumes
print_info "Backing up persistent volumes..."
kubectl get pvc -n kubernet-${ENVIRONMENT} -o json > "$BACKUP_DIR/pvc-backup.json"

# Backup configurations
print_info "Backing up configurations..."
kubectl get all -n kubernet-${ENVIRONMENT} -o yaml > "$BACKUP_DIR/k8s-resources.yaml"
helm get values auth-service -n kubernet-${ENVIRONMENT} > "$BACKUP_DIR/auth-service-values.yaml"

# Backup Vault secrets
print_info "Backing up Vault secrets..."
kubectl exec -n vault vault-0 -- vault operator raft snapshot save /tmp/raft-snapshot.db
kubectl cp vault/vault-0:/tmp/raft-snapshot.db "$BACKUP_DIR/vault-snapshot.db"

# Upload to S3
print_info "Uploading backup to S3..."
aws s3 cp "$BACKUP_DIR" s3://mini-kubernet-backups/${ENVIRONMENT}/${BACKUP_DATE}/ --recursive

print_info "Backup completed: $BACKUP_DIR"
print_info "S3 location: s3://mini-kubernet-backups/${ENVIRONMENT}/${BACKUP_DATE}/"

# Separate CRDs Implementation for KubeVuln

This implementation introduces two separate Custom Resource Definitions (CRDs) to replace the unified `scanreports` CRD:

- **SecretScanning CRD**: For trufflehog secret scanning reports
- **DiveScanning CRD**: For dive image layer analysis reports

## 🎯 Overview

Previously, both dive and trufflehog reports were stored in a single `scanreports` CRD. Now they are separated into dedicated CRDs for better organization and management.

## 📋 CRD Definitions

### 1. SecretScanning CRD (`secretscannings.kubevuln.io`)

**Purpose**: Stores trufflehog secret scanning results

**Key Fields**:
- `spec.image`: The scanned image
- `spec.trufflehogReport`: Complete JSON report from trufflehog
- `spec.secretsFound`: Number of secrets detected
- `spec.timestamp`: Scan timestamp
- `status.status`: Current status (completed, failed, etc.)

**Short Name**: `ss`

### 2. DiveScanning CRD (`divescannings.kubevuln.io`)

**Purpose**: Stores dive image layer analysis results

**Key Fields**:
- `spec.image`: The analyzed image
- `spec.diveReport`: Complete JSON report from dive
- `spec.imageSize`: Size of the analyzed image
- `spec.layers`: Number of image layers
- `spec.efficiencyScore`: Dive efficiency score
- `spec.timestamp`: Analysis timestamp
- `status.status`: Current status (completed, failed, etc.)

**Short Name**: `ds`

## 🚀 Deployment Instructions

### 1. Deploy the New CRDs

```bash
# Make the deployment script executable
chmod +x deploy-separate-crds.sh

# Deploy the CRDs to your EKS cluster
./deploy-separate-crds.sh
```

### 2. Verify CRD Installation

```bash
# Check if CRDs are installed
kubectl get crd | grep -E "(secretscannings|divescannings)"

# Verify CRD status
kubectl get crd secretscannings.kubevuln.io -o yaml
kubectl get crd divescannings.kubevuln.io -o yaml
```

### 3. Rebuild and Deploy KubeVuln

After deploying the CRDs, rebuild your kubevuln application with the new separate storage implementation:

```bash
# Build the application
make build

# Deploy to your EKS cluster
kubectl apply -f deploy/kubevuln-deployment.yaml
```

## 📊 Usage Examples

### View Secret Scanning Reports

```bash
# List all secret scanning reports
kubectl get secretscannings -A

# View reports in a specific namespace
kubectl get secretscannings -n cyberqshield

# Get detailed view of a specific report
kubectl get secretscannings <report-name> -n <namespace> -o yaml

# Watch for new secret scanning reports
kubectl get secretscannings -n <namespace> -w
```

### View Dive Scanning Reports

```bash
# List all dive scanning reports
kubectl get divescannings -A

# View reports in a specific namespace
kubectl get divescannings -n cyberqshield

# Get detailed view of a specific report
kubectl get divescannings <report-name> -n <namespace> -o yaml

# Watch for new dive scanning reports
kubectl get divescannings -n <namespace> -w
```

### Using Short Names

```bash
# Use short names for convenience
kubectl get ss -A  # Secret scanning reports
kubectl get ds -A  # Dive scanning reports
```

## 🔧 Code Changes Summary

### New Files Created:
1. `api/v1/secretscanning_types.go` - Go types for SecretScanning CRD
2. `api/v1/divescanning_types.go` - Go types for DiveScanning CRD
3. `adapters/v1/separate_crd_storage.go` - Storage adapter for separate CRDs
4. `deploy/secret-scanning-crd.yaml` - SecretScanning CRD definition
5. `deploy/dive-scanning-crd.yaml` - DiveScanning CRD definition
6. `deploy/separate-crds.yaml` - Combined CRD definitions file

### Modified Files:
1. `adapters/v1/dive.go` - Updated to use SeparateCRDStorageAdapter
2. `adapters/v1/trufflehog.go` - Updated to use SeparateCRDStorageAdapter
3. `adapters/v1/syft.go` - Updated to use SeparateCRDStorageAdapter
4. `adapters/v1/layer_analyzer.go` - Updated storage adapter
5. `cmd/http/main.go` - Updated to use new storage adapter
6. `test-kubevuln-automation.sh` - Updated to check new CRDs

## 🎨 Benefits

1. **Better Organization**: Separate CRDs for different scan types
2. **Improved Performance**: Dedicated schemas optimized for each scan type
3. **Enhanced Monitoring**: Easier to monitor specific scan types
4. **Cleaner Architecture**: Clear separation of concerns
5. **Better Querying**: More targeted kubectl queries

## 🏷️ CRD Labels and Annotations

### Secret Scanning Reports
- **Labels**:
  - `scanType: secret`
  - `tool: trufflehog`
  - `image: <sanitized-image-name>`
  - `jobId: <job-id>`

- **Annotations**:
  - `kubevuln.io/trufflehog-report: Complete trufflehog secrets scan JSON report`
  - `kubevuln.io/scan-tool: trufflehog`

### Dive Scanning Reports
- **Labels**:
  - `scanType: dive`
  - `tool: dive`
  - `image: <sanitized-image-name>`
  - `jobId: <job-id>`

- **Annotations**:
  - `kubevuln.io/dive-report: Complete dive analysis JSON report`
  - `kubevuln.io/scan-tool: dive`

## 🧪 Testing

Use the updated test automation script to verify the implementation:

```bash
# Run the automated test
./test-kubevuln-automation.sh
```

The script will now check for both SecretScanning and DiveScanning CRDs instead of the old unified scanreports CRD.

## 🔄 Migration from Old CRDs

If you have existing `scanreports` CRDs, you may want to:

1. Extract the data from existing reports
2. Clean up old CRDs (optional)
3. Use the new separate CRDs going forward

## 📚 Additional Information

- Both CRDs follow Kubernetes best practices
- Reports include complete JSON data from the respective tools
- Fallback to file storage is maintained for reliability
- CRDs are namespaced for proper multi-tenancy support

## 🛠️ Troubleshooting

### CRD Not Found
```bash
# Check if CRDs are properly installed
kubectl get crd | grep kubevuln.io
```

### No Reports Showing Up
```bash
# Check kubevuln pod logs
kubectl logs -f <kubevuln-pod-name> -n <namespace>

# Check if scans are running
kubectl exec <kubevuln-pod-name> -n <namespace> -- ls -la /tmp/
```

### Permission Issues
```bash
# Ensure proper RBAC permissions for the new CRDs
kubectl apply -f deploy/kubevuln-rbac.yaml
``` 
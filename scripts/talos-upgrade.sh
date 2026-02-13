#!/usr/bin/env bash

# Check if running with bash
if [ -z "$BASH_VERSION" ]; then
  echo "Error: This script requires bash. Please run with: bash $0"
  exit 1
fi

set -euo pipefail

# Talos Cluster Upgrade Script - System Extensions
# This script upgrades all Talos nodes to include system extensions for Longhorn and GPU support
# Date: February 8, 2026
# Schematic ID: 144f58860e456dda4f18038a2c7ebc91a4360f9a2b80458f03a6852f1ae12743

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCHEMATIC_ID="613e1592b2da41ae5e265e8789429f22e121aab91cb4deb6bc3c0b6262961245"  # iSCSI tools only
TALOS_VERSION="v1.9.3"
NEW_IMAGE="factory.talos.dev/installer/${SCHEMATIC_ID}:${TALOS_VERSION}"
TALOSCONFIG="${TALOSCONFIG:-./talosconfig}"
KUBECONFIG="${KUBECONFIG:-./kubeconfig}"

# Node IPs (from Pulumi stack output)
WORKERS=(
  "192.168.1.179:k8s-worker1"
  "192.168.1.172:k8s-worker2"
  "192.168.1.164:k8s-worker3"
  "192.168.1.87:k8s-worker4"
)

CONTROLLERS=(
  "192.168.1.247:k8s-controller1"
  "192.168.1.108:k8s-controller2"
  "192.168.1.216:k8s-controller3"
)

# Functions
log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Wait for node to be Ready with timeout
wait_for_node_ready() {
  local node_name=$1
  local timeout=${2:-300}  # Default 5 minutes
  
  log_info "Waiting for node ${node_name} to be Ready (timeout: ${timeout}s)..."
  
  if kubectl --kubeconfig="${KUBECONFIG}" wait --for=condition=Ready "node/${node_name}" --timeout="${timeout}s" 2>/dev/null; then
    log_success "Node ${node_name} is Ready"
    return 0
  else
    log_error "Node ${node_name} failed to become Ready within ${timeout}s"
    return 1
  fi
}

# Wait for Talos health check
wait_for_talos_health() {
  local node_ip=$1
  local max_attempts=60  # 5 minutes (5s intervals)
  local attempt=0
  
  log_info "Waiting for Talos health on ${node_ip}..."
  
  while [ $attempt -lt $max_attempts ]; do
    if talosctl --talosconfig="${TALOSCONFIG}" -n "${node_ip}" health --wait-timeout=5s &>/dev/null; then
      log_success "Talos health check passed for ${node_ip}"
      return 0
    fi
    
    attempt=$((attempt + 1))
    sleep 5
  done
  
  log_error "Talos health check failed for ${node_ip} after ${max_attempts} attempts"
  return 1
}

# Verify extensions installed
verify_extensions() {
  local node_ip=$1
  local expected_extensions=(
    "iscsi-tools"
    "util-linux-tools"
    "nvidia-container-toolkit"
    "nonfree-kmod-nvidia"
    "qemu-guest-agent"
  )
  
  log_info "Verifying extensions on ${node_ip}..."
  
  local extensions_output
  extensions_output=$(talosctl --talosconfig="${TALOSCONFIG}" -n "${node_ip}" get extensions 2>/dev/null || echo "")
  
  local all_found=true
  for ext in "${expected_extensions[@]}"; do
    if echo "${extensions_output}" | grep -q "${ext}"; then
      log_success "  ✓ ${ext}"
    else
      log_warning "  ✗ ${ext} not found"
      all_found=false
    fi
  done
  
  if [ "$all_found" = true ]; then
    log_success "All extensions verified on ${node_ip}"
    return 0
  else
    log_warning "Some extensions missing on ${node_ip}"
    return 1
  fi
}

# Verify iscsid service
verify_iscsid() {
  local node_ip=$1
  
  log_info "Verifying iscsid service on ${node_ip}..."
  
  if talosctl --talosconfig="${TALOSCONFIG}" -n "${node_ip}" service iscsid status 2>/dev/null | grep -q "STATE: Running"; then
    log_success "iscsid service is running on ${node_ip}"
    return 0
  else
    log_warning "iscsid service not running on ${node_ip}"
    return 1
  fi
}

# Upgrade a single node
upgrade_node() {
  local node_ip=$1
  local node_name=$2
  local node_type=$3  # "worker" or "controller"
  
  echo ""
  log_info "=========================================="
  log_info "Upgrading ${node_type}: ${node_name} (${node_ip})"
  log_info "=========================================="
  
  # Initiate upgrade
  log_info "Starting upgrade to image: ${NEW_IMAGE}"
  if ! talosctl --talosconfig="${TALOSCONFIG}" -n "${node_ip}" upgrade --image="${NEW_IMAGE}" --preserve; then
    log_error "Failed to initiate upgrade for ${node_name}"
    return 1
  fi
  
  log_success "Upgrade initiated for ${node_name}"
  
  # Wait for node to start rebooting (give it 10 seconds)
  log_info "Waiting for node to start rebooting..."
  sleep 10
  
  # Wait for Talos to be healthy
  if ! wait_for_talos_health "${node_ip}"; then
    log_error "Talos health check failed for ${node_name}"
    return 1
  fi
  
  # Wait for Kubernetes node to be Ready
  if ! wait_for_node_ready "${node_name}" 300; then
    log_error "Node ${node_name} failed to become Ready"
    return 1
  fi
  
  # Verify extensions (workers only)
  if [ "${node_type}" = "worker" ]; then
    if ! verify_extensions "${node_ip}"; then
      log_warning "Extension verification failed for ${node_name}, but continuing..."
    fi
    
    if ! verify_iscsid "${node_ip}"; then
      log_warning "iscsid verification failed for ${node_name}, but continuing..."
    fi
  fi
  
  # For controllers, verify etcd
  if [ "${node_type}" = "controller" ]; then
    log_info "Verifying etcd service on ${node_name}..."
    if talosctl --talosconfig="${TALOSCONFIG}" -n "${node_ip}" service etcd status 2>/dev/null | grep -q "STATE: Running"; then
      log_success "etcd service is running on ${node_name}"
    else
      log_warning "etcd service check inconclusive on ${node_name}"
    fi
  fi
  
  log_success "Successfully upgraded ${node_name}"
  return 0
}

# Pre-flight checks
preflight_checks() {
  log_info "Running pre-flight checks..."
  
  # Check talosconfig exists
  if [ ! -f "${TALOSCONFIG}" ]; then
    log_error "talosconfig not found at ${TALOSCONFIG}"
    log_info "Set TALOSCONFIG environment variable or run from infra/pulumi directory"
    exit 1
  fi
  
  # Check kubeconfig exists
  if [ ! -f "${KUBECONFIG}" ]; then
    log_error "kubeconfig not found at ${KUBECONFIG}"
    log_info "Set KUBECONFIG environment variable or run from infra/pulumi directory"
    exit 1
  fi
  
  # Check kubectl is available
  if ! command -v kubectl &> /dev/null; then
    log_error "kubectl not found in PATH"
    exit 1
  fi
  
  # Check talosctl is available
  if ! command -v talosctl &> /dev/null; then
    log_error "talosctl not found in PATH"
    exit 1
  fi
  
  # Check cluster connectivity
  log_info "Checking cluster connectivity..."
  if ! kubectl --kubeconfig="${KUBECONFIG}" get nodes &>/dev/null; then
    log_error "Cannot connect to Kubernetes cluster"
    exit 1
  fi
  
  log_success "Pre-flight checks passed"
}

# Backup cluster state
backup_cluster() {
  local backup_dir="./backups/talos-upgrade-$(date +%Y%m%d-%H%M%S)"
  
  log_info "Creating backup in ${backup_dir}..."
  mkdir -p "${backup_dir}"
  
  kubectl --kubeconfig="${KUBECONFIG}" get all --all-namespaces -o yaml > "${backup_dir}/all-resources.yaml" 2>/dev/null || true
  kubectl --kubeconfig="${KUBECONFIG}" get pv,pvc --all-namespaces -o yaml > "${backup_dir}/storage.yaml" 2>/dev/null || true
  kubectl --kubeconfig="${KUBECONFIG}" get nodes -o yaml > "${backup_dir}/nodes.yaml" 2>/dev/null || true
  
  log_success "Backup created at ${backup_dir}"
}

# Main upgrade process
main() {
  echo ""
  log_info "=========================================="
  log_info "Talos Cluster Upgrade - System Extensions"
  log_info "=========================================="
  log_info "Schematic ID: ${SCHEMATIC_ID}"
  log_info "Talos Version: ${TALOS_VERSION}"
  log_info "New Image: ${NEW_IMAGE}"
  echo ""
  
  # Pre-flight checks
  preflight_checks
  
  # Backup
  log_warning "Creating cluster backup before upgrade..."
  backup_cluster
  
  # Confirm before proceeding
  echo ""
  log_warning "This will upgrade all 7 nodes in the cluster."
  log_warning "Workers will be upgraded first, then control plane."
  log_warning "Each node will reboot during the upgrade."
  echo ""
  read -p "Do you want to proceed? (yes/no): " -r
  echo ""
  
  if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    log_info "Upgrade cancelled by user"
    exit 0
  fi
  
  # Phase 1: Upgrade Workers
  log_info "=========================================="
  log_info "PHASE 1: Upgrading Worker Nodes"
  log_info "=========================================="
  
  local failed_workers=()
  for worker_entry in "${WORKERS[@]}"; do
    IFS=':' read -r worker_ip worker_name <<< "${worker_entry}"
    
    if ! upgrade_node "${worker_ip}" "${worker_name}" "worker"; then
      log_error "Failed to upgrade ${worker_name}"
      failed_workers+=("${worker_name}")
    fi
    
    # Brief pause between workers
    sleep 5
  done
  
  # Check if any workers failed
  if [ ${#failed_workers[@]} -gt 0 ]; then
    log_error "The following workers failed to upgrade: ${failed_workers[*]}"
    log_error "Aborting control plane upgrade for safety"
    exit 1
  fi
  
  log_success "All workers upgraded successfully"
  
  # Phase 2: Upgrade Controllers
  log_info "=========================================="
  log_info "PHASE 2: Upgrading Control Plane Nodes"
  log_info "=========================================="
  
  local failed_controllers=()
  for controller_entry in "${CONTROLLERS[@]}"; do
    IFS=':' read -r controller_ip controller_name <<< "${controller_entry}"
    
    if ! upgrade_node "${controller_ip}" "${controller_name}" "controller"; then
      log_error "Failed to upgrade ${controller_name}"
      failed_controllers+=("${controller_name}")
    fi
    
    # Brief pause between controllers
    sleep 5
  done
  
  # Final status
  echo ""
  log_info "=========================================="
  log_info "UPGRADE COMPLETE"
  log_info "=========================================="
  
  if [ ${#failed_controllers[@]} -gt 0 ]; then
    log_error "The following controllers failed to upgrade: ${failed_controllers[*]}"
    exit 1
  fi
  
  log_success "All nodes upgraded successfully!"
  
  # Final verification
  echo ""
  log_info "Final cluster status:"
  kubectl --kubeconfig="${KUBECONFIG}" get nodes
  
  echo ""
  log_info "Next steps:"
  log_info "1. Verify Longhorn deployment: kubectl get pods -n longhorn-system"
  log_info "2. Check StorageClass: kubectl get storageclass"
  log_info "3. Verify Infisical PVCs: kubectl get pvc -n infisical"
  log_info "4. Monitor cluster for 24 hours"
}

# Run main function
main "$@"

#!/usr/bin/env bash
set -euo pipefail

# seal-secrets.sh
# Scans CDK8s output directory for Secret manifests and converts them to SealedSecrets
# Usage: ./seal-secrets.sh [dist-dir] [sealed-dir]

DIST_DIR="${1:-../../app}"
SEALED_DIR="${2:-./sealed}"
CONTROLLER_NAME="sealed-secrets-controller"
CONTROLLER_NAMESPACE="kube-system"

echo "🔍 Scanning for Secret manifests in ${DIST_DIR}..."

# Create sealed output directory
mkdir -p "${SEALED_DIR}"

# Create temp directory for processing (will be cleaned up)
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

# Find all YAML files in dist directory
find "${DIST_DIR}" -type f \( -name "*.yaml" -o -name "*.yml" \) | while read -r manifest_file; do
    echo "📄 Processing: ${manifest_file}"
    
    # Split multi-doc YAML into separate files
    # yq is used to parse YAML stream and extract Secret resources
    yq eval-all 'select(.kind == "Secret")' "${manifest_file}" > "${TMP_DIR}/secrets.yaml" 2>/dev/null || true
    
    # Check if any secrets were found
    if [ -s "${TMP_DIR}/secrets.yaml" ]; then
        # Count number of secrets
        secret_count=$(yq eval-all '. | length' "${TMP_DIR}/secrets.yaml")
        
        # Process each secret by index
        for ((i=0; i<secret_count; i++)); do
            # Extract individual secret
            yq eval-all "select(. | di == $i)" "${TMP_DIR}/secrets.yaml" > "${TMP_DIR}/secret-${i}.yaml"
            
            secret_name=$(yq eval '.metadata.name' "${TMP_DIR}/secret-${i}.yaml")
            secret_namespace=$(yq eval '.metadata.namespace // "default"' "${TMP_DIR}/secret-${i}.yaml")
            
            # Skip if secret name is null or empty
            if [ -z "${secret_name}" ] || [ "${secret_name}" = "null" ]; then
                echo "⚠️  Skipping secret with empty/null name in ${manifest_file}"
                rm -f "${TMP_DIR}/secret-${i}.yaml"
                continue
            fi
            
            echo "🔐 Sealing Secret: ${secret_namespace}/${secret_name}"
            
            # Seal the secret using kubeseal with local cert
            CERT_FILE="$(dirname "$0")/../sealed-secrets-cert.pem"
            
            if [ ! -f "${CERT_FILE}" ]; then
                echo "❌ ERROR: Certificate file not found: ${CERT_FILE}"
                echo "Please fetch it from your cluster:"
                echo "  kubeseal --fetch-cert --controller-name=sealed-secrets-controller --controller-namespace=kube-system > ${CERT_FILE}"
                exit 1
            fi
            
            kubeseal \
                --cert="${CERT_FILE}" \
                --format=yaml \
                --allow-empty-data \
                < "${TMP_DIR}/secret-${i}.yaml" \
                > "${SEALED_DIR}/sealed-${secret_namespace}-${secret_name}.yaml"
            
            echo "✅ Created: ${SEALED_DIR}/sealed-${secret_namespace}-${secret_name}.yaml"
            
            # Clean up this secret file
            rm -f "${TMP_DIR}/secret-${i}.yaml"
        done
    fi
    
    # Clean up temp files for this manifest
    rm -f "${TMP_DIR}"/secret-*.yml "${TMP_DIR}/secrets.yaml"
done

echo "✨ Sealing complete! SealedSecrets written to ${SEALED_DIR}/"

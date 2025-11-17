#!/usr/bin/env bash
set -e

TARGET_DIRECTORY=/etc/bifrost
ROOT_DOMAIN="madhan.app"
ADMIN_EMAIL="madhankumaravelu93@gmail.com"


make_absolute_path() {
    local path="$1"
    if [ "${path%/*}" = "$path" ]; then
        path="$(pwd)/$path"
    fi
    echo "$path"
}

wait_for_docker_ready() {
    echo "[INFO] Waiting for cloud-init and Docker daemon to become ready..."

    # Wait for cloud-init to finish if applicable
    while [ ! -f /var/lib/cloud/instance/boot-finished ]; do
        sleep 2
    done

    local retries=30
    while [ $retries -gt 0 ]; do
        if systemctl is-active --quiet docker; then
            if docker info >/dev/null 2>&1; then
                echo "[INFO] Docker is running and ready."
                return 0
            fi
        fi
        echo "[INFO] Docker not ready yet. Retrying in 5s..."
        sleep 5
        retries=$((retries - 1))
    done

    echo "[ERROR] Docker did not become ready in time."
    exit 1
}

generate_secret() {
    openssl rand -hex 32
}

docker_check() {
    if ! docker -v >/dev/null 2>&1; then
        echo "[ERROR] Failed to run 'docker -v'. Is Docker installed?"
        exit 1
    else
        echo "[INFO] Docker is installed."
    fi

    if ! docker image inspect hello-world:latest >/dev/null 2>&1; then
        echo "Pulling hello-world:latest image..."
        docker pull hello-world:latest >/dev/null 2>&1 || {
            echo "[ERROR] Failed to pull hello-world image." >&2
            exit 1
        }
    fi

    if docker run --rm hello-world:latest >/dev/null 2>&1; then
        echo "[INFO] Docker is working without root."
    else
        echo "[ERROR] Cannot run Docker containers without proper group permissions." >&2
        exit 1
    fi
}

jq_check() {
    if ! command -v jq >/dev/null 2>&1; then
        echo "[ERROR] 'jq' is not installed. Please install it before continuing."
        exit 1
    fi
}

create_files_and_directories() {
    echo "[INFO] Preparing environment files and directories..."

    mkdir -p data
    touch publisher.log

    keycloak_postgres_password=$(generate_secret)
    KEYCLOAK_ADMIN_PASSWORD=$(generate_secret)
    ZITI_ADMIN_PASSWORD=$(generate_secret)

    cat <<EOF > .env
# --- KEYCLOAK VARIABLES ---
KC_DB=postgres
KC_DB_URL=jdbc:postgresql://keycloak-postgres:5432/keycloak
KC_DB_PASSWORD=$keycloak_postgres_password
KC_DB_USERNAME=keycloak
KC_DB_SCHEMA=public
KEYCLOAK_ADMIN=admin
KEYCLOAK_ADMIN_PASSWORD=$KEYCLOAK_ADMIN_PASSWORD
PROXY_ADDRESS_FORWARDING=true
KC_METRICS_ENABLE=true
KC_PROXY=edge
KC_HOSTNAME_STRICT=false
KC_HOSTNAME_STRICT_HTTPS=false
KC_HTTP_ENABLED=true
KC_PROXY_HEADERS=xforwarded

# --- POSTGRES VARIABLES ---
POSTGRES_DB=keycloak
POSTGRES_USER=keycloak
POSTGRES_PASSWORD=$keycloak_postgres_password

# --- ZITI VARIABLES ---
ZITI_ADMIN_PASSWORD=$ZITI_ADMIN_PASSWORD
ZITI_CONTROLLER_URL=https://ziti.$ROOT_DOMAIN
ZITI_WEBSOCKET_CONTROLLER_URL=wss://ctrl.ziti.$ROOT_DOMAIN
ZITI_ADMIN_USERNAME=admin
EOF
}

install_ziti_cli() {
    echo "[INFO] Installing OpenZiti CLI..."
    if command -v ziti >/dev/null 2>&1; then
        return 0
    fi
    curl -sS https://get.openziti.io/install.bash | sudo bash -s openziti
}

ziti_login() {
    if [ -z "$ROOT_DOMAIN" ]; then 
        echo "[ERROR] Root domain is not specified."
        return 1
    fi
    ziti edge login "ctrl.ziti.$ROOT_DOMAIN:443" -u "admin" -p "${ZITI_ADMIN_PASSWORD}" -y 2>&1
}

install_ziti() {
    echo "[INFO] Installing and bootstrapping OpenZiti..."

    sed -i "s/ROOT_DOMAIN/$ROOT_DOMAIN/g" $TARGET_DIRECTORY/docker-compose.yml

    # bring up ziti-controller to generate the enrollment
    docker compose up -d
    echo "[INFO] Waiting 60s for ziti controller to start..."
    sleep 60

    ziti_login

    echo "[INFO] Creating router and enrollment token..."
    ziti edge delete edge-router "er1" >/dev/null 2>&1 || true
    ziti edge create edge-router "er1" -o "er1.jwt" -t -a "public"
    ROUTER_TOKEN=$(cat ./er1.jwt)
    rm -f er1.jwt

    echo "ZITI_ENROLL_TOKEN=$ROUTER_TOKEN" >> .env
    export $(grep -v '^#' .env | xargs)

    echo "[INFO] Restarting with enrollment token..."
    docker compose down
    docker compose up -d

    echo "[INFO] Waiting 60s for ziti router to enroll..."
    sleep 60
    docker compose down
}


main_install() {
    sudo echo "[INFO] Running setup as root user."
    wait_for_docker_ready
    docker_check
    jq_check
    create_files_and_directories
    install_ziti_cli
    install_ziti

    cat <<EOF

Setup complete! 🎉

echo "Run docker compose up -d to start the services."

EOF
}

main_install


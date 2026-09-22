#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
ENV_EXAMPLE="${ROOT_DIR}/.env.example"

echo "=========================================================="
echo "🛡️  AgentSight Preflight Port & Environment Verification"
echo "=========================================================="

# 1. Ensure .env exists
if [ ! -f "${ENV_FILE}" ]; then
    if [ -f "${ENV_EXAMPLE}" ]; then
        echo "📄 [PREFLIGHT] .env not found. Creating from .env.example..."
        cp "${ENV_EXAMPLE}" "${ENV_FILE}"
    else
        echo "📄 [PREFLIGHT] .env and .env.example not found. Creating default .env..."
        cat << 'EOF' > "${ENV_FILE}"
DB_USER=agentsight
DB_PASSWORD=agentsight
DB_NAME=agentsight
DB_PORT=5437
APP_PORT=8084
DATABASE_URL=postgres://agentsight:agentsight@127.0.0.1:5437/agentsight?sslmode=disable
SESSION_SECRET=agentsight-dev-secret-key-min-32-chars-long
BASE_URL=http://localhost:8084
APP_ENV=development
CHROME_BIN=/usr/bin/chromium-browser
EOF
    fi
fi

# Helper to read a key from .env
get_env_val() {
    local key="$1"
    local default_val="$2"
    local val
    val="$(grep -E "^${key}=" "${ENV_FILE}" | tail -n 1 | cut -d '=' -f 2- | tr -d '\r"' || true)"
    if [ -z "${val}" ]; then
        echo "${default_val}"
    else
        echo "${val}"
    fi
}

# Helper to update or append a key in .env
set_env_val() {
    local key="$1"
    local val="$2"
    if grep -q -E "^${key}=" "${ENV_FILE}"; then
        # Replace existing key safely
        sed -i -E "s|^${key}=.*|${key}=${val}|" "${ENV_FILE}"
    else
        # Append to end
        echo "${key}=${val}" >> "${ENV_FILE}"
    fi
}

# 2. Extract current port configurations
CURRENT_DB_PORT="$(get_env_val "DB_PORT" "5437")"
CURRENT_APP_PORT="$(get_env_val "APP_PORT" "8084")"

# 3. Port check utility
is_port_in_use() {
    local port="$1"
    # Check listening sockets using ss
    if ss -tulpn 2>/dev/null | grep -qE "(:${port}\s|:${port}$)"; then
        return 0
    fi
    # Check using lsof if available
    if command -v lsof >/dev/null 2>&1; then
        if lsof -iTCP:"${port}" -sTCP:LISTEN -P -n >/dev/null 2>&1; then
            return 0
        fi
    fi
    # Check using netcat if available
    if command -v nc >/dev/null 2>&1; then
        if nc -z 127.0.0.1 "${port}" 2>/dev/null; then
            return 0
        fi
    fi
    return 1
}

find_free_port() {
    local candidate="$1"
    local name="$2"
    while is_port_in_use "${candidate}"; do
        echo "⚠️  [PREFLIGHT] ${name} port ${candidate} is occupied. Auto-incrementing..." >&2
        candidate=$((candidate + 1))
    done
    echo "${candidate}"
}

# 4. Resolve free ports
FREE_DB_PORT="$(find_free_port "${CURRENT_DB_PORT}" "PostgreSQL (DB_PORT)")"
FREE_APP_PORT="$(find_free_port "${CURRENT_APP_PORT}" "Application (APP_PORT)")"

# 5. Update .env if ports were modified or needed
CHANGED=0

if [ "${FREE_DB_PORT}" != "${CURRENT_DB_PORT}" ]; then
    echo "🔄 [PREFLIGHT] Updating DB_PORT to ${FREE_DB_PORT} in .env..."
    set_env_val "DB_PORT" "${FREE_DB_PORT}"
    # Update DATABASE_URL host port if present
    DB_USER="$(get_env_val "DB_USER" "agentsight")"
    DB_PASSWORD="$(get_env_val "DB_PASSWORD" "agentsight")"
    DB_NAME="$(get_env_val "DB_NAME" "agentsight")"
    set_env_val "DATABASE_URL" "postgres://${DB_USER}:${DB_PASSWORD}@127.0.0.1:${FREE_DB_PORT}/${DB_NAME}?sslmode=disable"
    CHANGED=1
else
    echo "✅ [PREFLIGHT] DB_PORT=${FREE_DB_PORT} is free."
fi

if [ "${FREE_APP_PORT}" != "${CURRENT_APP_PORT}" ]; then
    echo "🔄 [PREFLIGHT] Updating APP_PORT to ${FREE_APP_PORT} in .env..."
    set_env_val "APP_PORT" "${FREE_APP_PORT}"
    set_env_val "BASE_URL" "http://localhost:${FREE_APP_PORT}"
    CHANGED=1
else
    echo "✅ [PREFLIGHT] APP_PORT=${FREE_APP_PORT} is free."
fi

# Ensure all critical keys are present in .env
ensure_key() {
    local key="$1"
    local default_val="$2"
    if ! grep -q -E "^${key}=" "${ENV_FILE}"; then
        set_env_val "${key}" "${default_val}"
    fi
}

ensure_key "DB_USER" "agentsight"
ensure_key "DB_PASSWORD" "agentsight"
ensure_key "DB_NAME" "agentsight"
ensure_key "DB_PORT" "${FREE_DB_PORT}"
ensure_key "APP_PORT" "${FREE_APP_PORT}"
ensure_key "SESSION_SECRET" "change-me-in-production-min-32-chars"
ensure_key "CHROME_BIN" "/usr/bin/chromium-browser"

echo ""
echo "🚀 [PREFLIGHT] Configuration ready:"
echo "   - DB_PORT:  ${FREE_DB_PORT}"
echo "   - APP_PORT: ${FREE_APP_PORT}"
echo "   - .env:     ${ENV_FILE}"
echo "=========================================================="

exit 0

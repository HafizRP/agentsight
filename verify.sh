#!/usr/bin/env bash
set -euo pipefail

# Determine APP_PORT from environment, .env file, or default to 8084
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "${SCRIPT_DIR}/.env" ]; then
    ENV_PORT="$(grep -E "^APP_PORT=" "${SCRIPT_DIR}/.env" | tail -n 1 | cut -d '=' -f 2- | tr -d '\r"' || true)"
    if [ -n "${ENV_PORT}" ]; then
        APP_PORT="${ENV_PORT}"
    fi
fi
APP_PORT="${APP_PORT:-8084}"
BASE_URL="http://127.0.0.1:${APP_PORT}"

echo "=========================================================="
echo "🧪 AgentSight Automated Smoke Test Suite"
echo "Target Base URL: ${BASE_URL}"
echo "=========================================================="

PASS=0
FAIL=0

run_test() {
    local test_num="$1"
    local test_name="$2"
    shift 2

    echo -n "Test ${test_num}: ${test_name} ... "
    if "$@"; then
        echo "✅ PASS"
        PASS=$((PASS + 1))
    else
        echo "❌ FAIL"
        FAIL=$((FAIL + 1))
    fi
}

# Test 1: Health check — GET /healthz → 200
test_healthz() {
    local code
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/healthz")"
    [ "${code}" = "200" ]
}

# Test 2: Readiness check — GET /readyz → 200
test_readyz() {
    local code
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/readyz")"
    [ "${code}" = "200" ]
}

# Test 3: Landing page — GET / → 200, contains "AgentSight" in body
test_landing() {
    local code body
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/")"
    body="$(curl -s "${BASE_URL}/")"
    [ "${code}" = "200" ] && [[ "${body}" == *"AgentSight"* ]]
}

# Test 4: Search — GET /search?q=react → 200, contains skill results
test_search() {
    local code body
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/search?q=react")"
    body="$(curl -s "${BASE_URL}/search?q=react")"
    [ "${code}" = "200" ] && ([[ "${body}" == *"skill-card"* ]] || [[ "${body}" == *"/skill/"* ]])
}

# Test 5: Trending — GET /trending → 200, contains skill cards
test_trending() {
    local code body
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/trending")"
    body="$(curl -s "${BASE_URL}/trending")"
    [ "${code}" = "200" ] && [[ "${body}" == *"/skill/"* ]]
}

# Test 6: Browse by platform — GET /platform/cursor → 200
test_platform() {
    local code body
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/platform/cursor")"
    body="$(curl -s "${BASE_URL}/platform/cursor")"
    [ "${code}" = "200" ] && [[ "${body}" == *"cursor"* ]]
}

# Test 7: Browse by category — GET /category/rules → 200
test_category() {
    local code body
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/category/rules")"
    body="$(curl -s "${BASE_URL}/category/rules")"
    [ "${code}" = "200" ] && [[ "${body}" == *"rules"* ]]
}

# Test 8: API skills list — GET /api/v1/skills → 200, valid JSON array
test_api_skills() {
    local code res
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/skills")"
    res="$(curl -s "${BASE_URL}/api/v1/skills")"
    [ "${code}" = "200" ] && echo "${res}" | jq -e '.skills | type == "array"' >/dev/null
}

# Test 9: API search — GET /api/v1/search?q=mcp → 200, valid JSON
test_api_search() {
    local code res
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/search?q=mcp")"
    res="$(curl -s "${BASE_URL}/api/v1/search?q=mcp")"
    [ "${code}" = "200" ] && echo "${res}" | jq -e '.skills | type == "array"' >/dev/null
}

# Test 10: API trending — GET /api/v1/trending → 200, valid JSON
test_api_trending() {
    local code res
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/trending")"
    res="$(curl -s "${BASE_URL}/api/v1/trending")"
    [ "${code}" = "200" ] && echo "${res}" | jq -e '.skills | type == "array"' >/dev/null
}

# Test 11: API stats — GET /api/v1/stats → 200, JSON with platform counts
test_api_stats() {
    local code res
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/stats")"
    res="$(curl -s "${BASE_URL}/api/v1/stats")"
    [ "${code}" = "200" ] && echo "${res}" | jq -e '(.platform_counts | type == "object") and (.platform_counts | length > 0)' >/dev/null
}

# Test 12: Skill detail — GET /skill/{first-slug-from-api} → 200
test_skill_detail() {
    local first_slug code
    first_slug="$(curl -s "${BASE_URL}/api/v1/skills" | jq -r '.skills[0].slug // empty')"
    [ -n "${first_slug}" ] || return 1
    code="$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/skill/${first_slug}")"
    [ "${code}" = "200" ]
}

# Test 13: Port security — DB port bound to 127.0.0.1, NOT 0.0.0.0
test_port_security() {
    local binding
    binding="$(docker port agentsight_postgres 5432 2>/dev/null || ss -tulpn | grep 5437 || true)"
    [ -n "${binding}" ] || return 1
    echo "${binding}" | grep -q "127.0.0.1" && ! echo "${binding}" | grep -q "0.0.0.0"
}

# Test 14: Seed data check — API returns >= 50 skills total
test_seed_data() {
    local total
    total="$(curl -s "${BASE_URL}/api/v1/skills" | jq -r '.total // 0')"
    [ "${total}" -ge 50 ]
}

# Execute tests
run_test "1"  "Health check (GET /healthz)" test_healthz
run_test "2"  "Readiness check (GET /readyz)" test_readyz
run_test "3"  "Landing page (GET /)" test_landing
run_test "4"  "Search page (GET /search?q=react)" test_search
run_test "5"  "Trending page (GET /trending)" test_trending
run_test "6"  "Browse by platform (GET /platform/cursor)" test_platform
run_test "7"  "Browse by category (GET /category/rules)" test_category
run_test "8"  "API skills list (GET /api/v1/skills)" test_api_skills
run_test "9"  "API search (GET /api/v1/search?q=mcp)" test_api_search
run_test "10" "API trending (GET /api/v1/trending)" test_api_trending
run_test "11" "API stats (GET /api/v1/stats)" test_api_stats
run_test "12" "Skill detail (GET /skill/{first-slug})" test_skill_detail
run_test "13" "Port security (DB port bound to 127.0.0.1 only)" test_port_security
run_test "14" "Seed data check (>= 50 skills indexed)" test_seed_data

echo "=========================================================="
echo "Results: ${PASS} passed, ${FAIL} failed"
echo "=========================================================="

if [ "${FAIL}" -eq 0 ]; then
    echo "🎉 All tests passed successfully!"
    exit 0
else
    echo "⚠️ One or more tests failed."
    exit 1
fi

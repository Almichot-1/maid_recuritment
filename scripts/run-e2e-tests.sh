#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
OUT_DIR="${OUT_DIR:-${ROOT_DIR}/reports}"
mkdir -p "${OUT_DIR}"

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
API_LOG="${OUT_DIR}/api-${TIMESTAMP}.log"
GO_TEST_LOG="${OUT_DIR}/go-test-${TIMESTAMP}.log"
API_E2E_LOG="${OUT_DIR}/api-e2e-${TIMESTAMP}.log"
REPORT_FILE="${OUT_DIR}/e2e-report-${TIMESTAMP}.md"

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required tool: $1"
    exit 1
  fi
}

require_tool curl
require_tool jq
require_tool go

cleanup() {
  if [[ -n "${API_PID:-}" ]] && kill -0 "${API_PID}" 2>/dev/null; then
    kill "${API_PID}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

health_check() {
  curl -sS -f "${API_BASE_URL}/health" >/dev/null 2>&1
}

start_api_if_needed() {
  if health_check; then
    echo "API already running at ${API_BASE_URL}"
    return
  fi

  echo "Starting API server..."
  go run ./cmd/api >"${API_LOG}" 2>&1 &
  API_PID=$!

  local i
  for i in {1..60}; do
    if health_check; then
      echo "API is healthy"
      return
    fi
    sleep 1
  done

  echo "API failed to become healthy. Check ${API_LOG}"
  exit 1
}

api_call_json() {
  local method="$1"; shift
  local url="$1"; shift
  local body="${1:-}"
  local token="${2:-}"

  local headers=(-H "Content-Type: application/json")
  if [[ -n "${token}" ]]; then
    headers+=(-H "Authorization: Bearer ${token}")
  fi

  local response code payload
  if [[ -n "${body}" ]]; then
    response=$(curl -sS -w "\n%{http_code}" -X "${method}" "${url}" "${headers[@]}" -d "${body}")
  else
    response=$(curl -sS -w "\n%{http_code}" -X "${method}" "${url}" "${headers[@]}")
  fi

  code=$(echo "${response}" | tail -n1)
  payload=$(echo "${response}" | sed '$d')

  if [[ "${code}" -lt 200 || "${code}" -ge 300 ]]; then
    echo "API call failed [${method} ${url}] HTTP ${code}" | tee -a "${API_E2E_LOG}"
    echo "${payload}" | tee -a "${API_E2E_LOG}"
    return 1
  fi

  echo "${payload}"
}

run_api_smoke_suite() {
  source "${OUT_DIR}/seed-output.env"

  echo "Running API E2E smoke suite" | tee -a "${API_E2E_LOG}"

  local selection_resp selection_id candidate_resp candidate_status approvals_resp

  selection_resp=$(api_call_json POST "${API_BASE_URL}/candidates/${CANDIDATE_1_ID}/select" "" "${FOR_TOKEN}")
  selection_id=$(echo "${selection_resp}" | jq -r '.selection.id')
  [[ -n "${selection_id}" && "${selection_id}" != "null" ]] || {
    echo "Missing selection id in response" | tee -a "${API_E2E_LOG}"
    return 1
  }

  candidate_resp=$(api_call_json GET "${API_BASE_URL}/candidates/${CANDIDATE_1_ID}" "" "${ETH_TOKEN}")
  candidate_status=$(echo "${candidate_resp}" | jq -r '.candidate.status')
  [[ "${candidate_status}" == "locked" ]] || {
    echo "Expected candidate to be locked after selection, got: ${candidate_status}" | tee -a "${API_E2E_LOG}"
    return 1
  }

  api_call_json POST "${API_BASE_URL}/selections/${selection_id}/approve" "" "${ETH_TOKEN}" >/dev/null
  approvals_resp=$(api_call_json POST "${API_BASE_URL}/selections/${selection_id}/approve" "" "${FOR_TOKEN}")

  local is_fully_approved
  is_fully_approved=$(echo "${approvals_resp}" | jq -r '.is_fully_approved // false')
  if [[ "${is_fully_approved}" != "true" ]]; then
    echo "Selection is not fully approved yet" | tee -a "${API_E2E_LOG}"
  fi

  api_call_json GET "${API_BASE_URL}/candidates/${CANDIDATE_1_ID}/status-steps" "" "${ETH_TOKEN}" >/dev/null
  api_call_json GET "${API_BASE_URL}/notifications/" "" "${ETH_TOKEN}" >/dev/null
  api_call_json GET "${API_BASE_URL}/notifications/" "" "${FOR_TOKEN}" >/dev/null

  echo "API E2E smoke suite passed" | tee -a "${API_E2E_LOG}"
}

generate_report() {
  local api_result="$1"
  local go_test_result="$2"

  cat > "${REPORT_FILE}" <<EOF
# E2E Test Report

- Timestamp: ${TIMESTAMP}
- API Base URL: ${API_BASE_URL}
- API smoke result: ${api_result}
- Go test suite result: ${go_test_result}

## Artifacts

- API server log: ${API_LOG}
- API smoke log: ${API_E2E_LOG}
- Go test log: ${GO_TEST_LOG}
- Seed output: ${OUT_DIR}/seed-output.env

## Commands Executed

1. Started API server (if needed)
2. scripts/seed-test-data.sh
3. API smoke workflow (select -> lock verify -> dual approve -> status/notifications checks)
4. go test ./...
EOF

  echo "Report written: ${REPORT_FILE}"
}

main() {
  start_api_if_needed

  echo "Seeding test data..."
  API_BASE_URL="${API_BASE_URL}" OUT_DIR="${OUT_DIR}" bash "${ROOT_DIR}/scripts/seed-test-data.sh"

  local api_result="PASS"
  local go_test_result="PASS"

  if ! run_api_smoke_suite; then
    api_result="FAIL"
  fi

  echo "Running Go test suite..."
  if ! go test ./... >"${GO_TEST_LOG}" 2>&1; then
    go_test_result="FAIL"
  fi

  generate_report "${api_result}" "${go_test_result}"

  if [[ "${api_result}" == "FAIL" || "${go_test_result}" == "FAIL" ]]; then
    echo "E2E run completed with failures. See report: ${REPORT_FILE}"
    exit 1
  fi

  echo "E2E run completed successfully. See report: ${REPORT_FILE}"
}

main "$@"

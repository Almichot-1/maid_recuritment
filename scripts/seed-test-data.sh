#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
ET_EMAIL="${ET_EMAIL:-ethiopian.e2e+$(date +%s)@example.com}"
FOR_EMAIL="${FOR_EMAIL:-foreign.e2e+$(date +%s)@example.com}"
TEST_PASSWORD="${TEST_PASSWORD:-Password123!}"

PASSPORT_URL="${PASSPORT_URL:-https://example.com/test/passport.pdf}"
PHOTO_URL="${PHOTO_URL:-https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d}"
VIDEO_URL="${VIDEO_URL:-https://samplelib.com/lib/preview/mp4/sample-5s.mp4}"

OUT_DIR="${OUT_DIR:-./reports}"
mkdir -p "${OUT_DIR}"

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required tool: $1"
    exit 1
  fi
}

require_tool curl
require_tool jq

api_json() {
  local method="$1"; shift
  local url="$1"; shift
  local body="${1:-}"
  local auth="${2:-}"

  local headers=(-H "Content-Type: application/json")
  if [[ -n "${auth}" ]]; then
    headers+=(-H "Authorization: Bearer ${auth}")
  fi

  local resp code
  if [[ -n "${body}" ]]; then
    resp=$(curl -sS -w "\n%{http_code}" -X "${method}" "${url}" "${headers[@]}" -d "${body}")
  else
    resp=$(curl -sS -w "\n%{http_code}" -X "${method}" "${url}" "${headers[@]}")
  fi

  code=$(echo "${resp}" | tail -n1)
  body=$(echo "${resp}" | sed '$d')

  echo "${body}" > "${OUT_DIR}/last-response.json"

  if [[ "${code}" -lt 200 || "${code}" -ge 300 ]]; then
    echo "Request failed: ${method} ${url} (HTTP ${code})"
    cat "${OUT_DIR}/last-response.json"
    return 1
  fi

  cat "${OUT_DIR}/last-response.json"
}

register_user() {
  local email="$1"
  local full_name="$2"
  local role="$3"
  local company_name="$4"

  local payload
  payload=$(jq -n \
    --arg email "${email}" \
    --arg password "${TEST_PASSWORD}" \
    --arg full_name "${full_name}" \
    --arg role "${role}" \
    --arg company_name "${company_name}" \
    '{email:$email,password:$password,full_name:$full_name,role:$role,company_name:$company_name}')

  local resp code
  resp=$(curl -sS -w "\n%{http_code}" -X POST "${API_BASE_URL}/auth/register" -H "Content-Type: application/json" -d "${payload}")
  code=$(echo "${resp}" | tail -n1)

  if [[ "${code}" == "201" ]]; then
    echo "Registered ${role}: ${email}"
    return 0
  fi

  if [[ "${code}" == "409" ]]; then
    echo "User exists, continuing: ${email}"
    return 0
  fi

  echo "Registration failed for ${email} (HTTP ${code})"
  echo "${resp}" | sed '$d'
  return 1
}

login_get_token() {
  local email="$1"
  local payload
  payload=$(jq -n --arg email "${email}" --arg password "${TEST_PASSWORD}" '{email:$email,password:$password}')

  api_json POST "${API_BASE_URL}/auth/login" "${payload}" | jq -r '.token'
}

create_candidate() {
  local et_token="$1"
  local name="$2"
  local age="$3"
  local exp="$4"

  local payload
  payload=$(jq -n \
    --arg full_name "${name}" \
    --argjson age "${age}" \
    --argjson experience_years "${exp}" \
    '{full_name:$full_name,age:$age,experience_years:$experience_years,languages:["English","Amharic"],skills:["Cooking","Cleaning"]}')

  api_json POST "${API_BASE_URL}/candidates/" "${payload}" "${et_token}" | jq -r '.candidate.id'
}

upload_document() {
  local et_token="$1"
  local candidate_id="$2"
  local doc_type="$3"
  local file_url="$4"
  local file_name="$5"

  local payload
  payload=$(jq -n \
    --arg document_type "${doc_type}" \
    --arg file_url "${file_url}" \
    --arg file_name "${file_name}" \
    --argjson file_size 1024 \
    '{document_type:$document_type,file_url:$file_url,file_name:$file_name,file_size:$file_size}')

  api_json POST "${API_BASE_URL}/candidates/${candidate_id}/documents" "${payload}" "${et_token}" >/dev/null
}

publish_candidate() {
  local et_token="$1"
  local candidate_id="$2"
  api_json POST "${API_BASE_URL}/candidates/${candidate_id}/publish" "" "${et_token}" >/dev/null
}

generate_cv() {
  local et_token="$1"
  local candidate_id="$2"
  api_json POST "${API_BASE_URL}/candidates/${candidate_id}/generate-cv" "" "${et_token}" >/dev/null
}

echo "Seeding test users and candidates against ${API_BASE_URL}"

register_user "${ET_EMAIL}" "E2E Ethiopian Agent" "ethiopian_agent" ""
register_user "${FOR_EMAIL}" "E2E Foreign Agent" "foreign_agent" "Global Agency"

ETH_TOKEN=$(login_get_token "${ET_EMAIL}")
FOR_TOKEN=$(login_get_token "${FOR_EMAIL}")

if [[ -z "${ETH_TOKEN}" || "${ETH_TOKEN}" == "null" ]]; then
  echo "Failed to get Ethiopian token"
  exit 1
fi
if [[ -z "${FOR_TOKEN}" || "${FOR_TOKEN}" == "null" ]]; then
  echo "Failed to get Foreign token"
  exit 1
fi

CANDIDATE_1_ID=$(create_candidate "${ETH_TOKEN}" "E2E Candidate One" 24 3)
CANDIDATE_2_ID=$(create_candidate "${ETH_TOKEN}" "E2E Candidate Two" 29 5)

upload_document "${ETH_TOKEN}" "${CANDIDATE_1_ID}" "passport" "${PASSPORT_URL}" "passport.pdf"
upload_document "${ETH_TOKEN}" "${CANDIDATE_1_ID}" "photo" "${PHOTO_URL}" "photo.jpg"
upload_document "${ETH_TOKEN}" "${CANDIDATE_1_ID}" "video" "${VIDEO_URL}" "video.mp4"

generate_cv "${ETH_TOKEN}" "${CANDIDATE_1_ID}"
publish_candidate "${ETH_TOKEN}" "${CANDIDATE_1_ID}"
publish_candidate "${ETH_TOKEN}" "${CANDIDATE_2_ID}"

cat > "${OUT_DIR}/seed-output.env" <<EOF
API_BASE_URL=${API_BASE_URL}
ET_EMAIL=${ET_EMAIL}
FOR_EMAIL=${FOR_EMAIL}
ETH_TOKEN=${ETH_TOKEN}
FOR_TOKEN=${FOR_TOKEN}
CANDIDATE_1_ID=${CANDIDATE_1_ID}
CANDIDATE_2_ID=${CANDIDATE_2_ID}
EOF

echo "Seed complete. Output: ${OUT_DIR}/seed-output.env"

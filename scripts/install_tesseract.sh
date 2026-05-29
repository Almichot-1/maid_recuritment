#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="${ROOT_DIR}/.render/tesseract"
TESSDATA_DIR="${INSTALL_DIR}/tessdata"

mkdir -p "${TESSDATA_DIR}"

TESSERACT_URL="https://github.com/DanielMYT/tesseract-static/releases/download/tesseract-5.5.2/tesseract.x86_64"
ENG_DATA_URL="https://raw.githubusercontent.com/tesseract-ocr/tessdata_fast/main/eng.traineddata"
OCRB_DATA_URL="https://raw.githubusercontent.com/Shreeshrii/tessdata_ocrb/master/ocrb.traineddata"

SYSTEM_TESSERACT=""

if command -v apt-get >/dev/null 2>&1; then
  export DEBIAN_FRONTEND=noninteractive
  if apt-get update && apt-get install -y --no-install-recommends tesseract-ocr tesseract-ocr-eng >/dev/null 2>&1; then
    if command -v tesseract >/dev/null 2>&1; then
      SYSTEM_TESSERACT="$(command -v tesseract)"
    fi
  fi
fi

if [ -n "${SYSTEM_TESSERACT}" ]; then
  ln -sf "${SYSTEM_TESSERACT}" "${INSTALL_DIR}/tesseract"
  echo "Using system Tesseract at ${SYSTEM_TESSERACT}"
else
  curl --fail --location --retry 3 --output "${INSTALL_DIR}/tesseract" "${TESSERACT_URL}"
  chmod +x "${INSTALL_DIR}/tesseract"
fi

curl --fail --location --retry 3 --output "${TESSDATA_DIR}/eng.traineddata" "${ENG_DATA_URL}"
curl --fail --location --retry 3 --output "${TESSDATA_DIR}/ocrb.traineddata" "${OCRB_DATA_URL}"

if [ -n "${SYSTEM_TESSERACT}" ]; then
  echo "Prepared tessdata for system Tesseract in ${INSTALL_DIR}"
else
  echo "Installed static Tesseract to ${INSTALL_DIR}"
fi

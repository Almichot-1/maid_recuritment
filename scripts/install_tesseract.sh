#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="${ROOT_DIR}/.render/tesseract"
TESSDATA_DIR="${INSTALL_DIR}/tessdata"

mkdir -p "${TESSDATA_DIR}"

TESSERACT_URL="https://github.com/DanielMYT/tesseract-static/releases/download/tesseract-5.5.2/tesseract.x86_64"
ENG_DATA_URL="https://raw.githubusercontent.com/tesseract-ocr/tessdata_fast/main/eng.traineddata"
OCRB_DATA_URL="https://raw.githubusercontent.com/Shreeshrii/tessdata_ocrb/master/ocrb.traineddata"

curl --fail --location --retry 3 --output "${INSTALL_DIR}/tesseract" "${TESSERACT_URL}"
chmod +x "${INSTALL_DIR}/tesseract"

curl --fail --location --retry 3 --output "${TESSDATA_DIR}/eng.traineddata" "${ENG_DATA_URL}"
curl --fail --location --retry 3 --output "${TESSDATA_DIR}/ocrb.traineddata" "${OCRB_DATA_URL}"

echo "Installed static Tesseract to ${INSTALL_DIR}"

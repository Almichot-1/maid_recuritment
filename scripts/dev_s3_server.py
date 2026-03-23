from __future__ import annotations

import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlparse


ROOT = Path(os.environ.get("DEV_S3_ROOT", r"C:\Users\NOOR AL MUSABAH\Documents\PROJECT_2\.dev-s3-data"))
ROOT.mkdir(parents=True, exist_ok=True)


CONTENT_TYPES = {
    ".pdf": "application/pdf",
    ".png": "image/png",
    ".jpg": "image/jpeg",
    ".jpeg": "image/jpeg",
    ".mp4": "video/mp4",
}


class DevS3Handler(BaseHTTPRequestHandler):
    server_version = "DevS3/0.1"

    def do_PUT(self) -> None:
        target = self._target_path()
        target.parent.mkdir(parents=True, exist_ok=True)
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length)
        target.write_bytes(body)
        self.send_response(200)
        self.end_headers()

    def do_DELETE(self) -> None:
        target = self._target_path()
        if target.exists():
            target.unlink()
        self.send_response(204)
        self.end_headers()

    def do_GET(self) -> None:
        target = self._target_path()
        if not target.exists() or not target.is_file():
            self.send_response(404)
            self.end_headers()
            return

        suffix = target.suffix.lower()
        content_type = CONTENT_TYPES.get(suffix, "application/octet-stream")
        body = target.read_bytes()

        self.send_response(200)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format: str, *args: object) -> None:
        return

    def _target_path(self) -> Path:
        parsed = urlparse(self.path)
        cleaned = parsed.path.lstrip("/")
        safe_parts = [part for part in cleaned.split("/") if part not in {"", ".", ".."}]
        return ROOT.joinpath(*safe_parts)


def main() -> None:
    host = os.environ.get("DEV_S3_HOST", "127.0.0.1")
    port = int(os.environ.get("DEV_S3_PORT", "9000"))
    server = ThreadingHTTPServer((host, port), DevS3Handler)
    server.serve_forever()


if __name__ == "__main__":
    main()

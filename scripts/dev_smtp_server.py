from __future__ import annotations

import asyncio
import os
from pathlib import Path


LOG_PATH = Path(os.environ.get("DEV_SMTP_LOG", r"C:\Users\NOOR AL MUSABAH\Documents\PROJECT_2\smtp-live.log"))


async def handle_client(reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
    def write_line(line: str) -> None:
        writer.write((line + "\r\n").encode())

    write_line("220 localhost Dev SMTP")
    await writer.drain()

    in_data = False
    message_lines: list[str] = []

    while not reader.at_eof():
        raw = await reader.readline()
        if not raw:
            break

        line = raw.decode(errors="replace").rstrip("\r\n")
        upper = line.upper()

        if in_data:
            if line == ".":
                LOG_PATH.parent.mkdir(parents=True, exist_ok=True)
                with LOG_PATH.open("a", encoding="utf-8") as handle:
                    handle.write("\n--- MESSAGE ---\n")
                    handle.write("\n".join(message_lines))
                    handle.write("\n")
                message_lines = []
                in_data = False
                write_line("250 OK")
            else:
                message_lines.append(line)
            await writer.drain()
            continue

        if upper.startswith("EHLO") or upper.startswith("HELO"):
            write_line("250-localhost")
            write_line("250-AUTH PLAIN")
            write_line("250 OK")
        elif upper.startswith("AUTH "):
            write_line("235 Authentication successful")
        elif upper.startswith("MAIL FROM") or upper.startswith("RCPT TO"):
            write_line("250 OK")
        elif upper == "DATA":
            in_data = True
            write_line("354 End data with <CR><LF>.<CR><LF>")
        elif upper == "QUIT":
            write_line("221 Bye")
            await writer.drain()
            break
        else:
            write_line("250 OK")

        await writer.drain()

    writer.close()
    await writer.wait_closed()


async def main() -> None:
    host = os.environ.get("DEV_SMTP_HOST", "127.0.0.1")
    port = int(os.environ.get("DEV_SMTP_PORT", "1025"))
    server = await asyncio.start_server(handle_client, host, port)
    async with server:
        await server.serve_forever()


if __name__ == "__main__":
    asyncio.run(main())

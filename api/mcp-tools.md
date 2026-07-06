# MCP Tools

Dokumentasi tool MCP yang di-expose oleh `internal/adapter/mcp/`.
Setiap tool: nama, deskripsi, parameter, dan apakah ReadOnly/Destructive
(lihat TECH-STACK-DAN-PERSIAPAN.md §5 soal ToolAnnotations).

| Tool | ReadOnly/Destructive | Deskripsi |
|---|---|---|
| `get_device_status` | ReadOnly | Cek status device (via Translate + Execute) |
| `run_command` | Bergantung command policy | Eksekusi command mentah |
| `push_config` | Destructive | Push konfigurasi ke device |

> ⚠ Re-review setelah 28 Juli 2026: revisi stateless MCP spec bisa
> mengubah asumsi handshake/session di adapter ini. Lihat
> `TECH-STACK-DAN-PERSIAPAN.md` bagian pembuka.

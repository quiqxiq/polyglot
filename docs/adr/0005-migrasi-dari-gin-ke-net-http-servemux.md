# 0005 — Migrasi dari Gin ke net/http.ServeMux Standar Go 1.22+

## Status
Diterima (menggantikan / supersedes [0001-pilih-gin-daripada-echo.md](file:///c:/Users/g0str/projects/polyground/polyglot/docs/adr/0001-pilih-gin-daripada-echo.md))

## Konteks
Pada fase awal proyek (ADR 0001), Gin dipilih sebagai web framework. Namun seiring evolusi arsitektur ke arah ConnectRPC (`connectrpc.com/connect`) dan Go 1.22+ yang sudah mendukung pattern routing bawaan (`METHOD /path/{param}`), Gin hanya berfungsi sebagai wrapper tipis (`gin.WrapH`) di atas handler native `net/http.Handler`.

Penggunaan Gin sebagai wrapper menambahkan layer overhead yang tidak perlu, memaksa adapter WebSocket/SSE bergantung pada context Gin yang kaku, serta memperumit middleware chain.

## Keputusan
1. Menghapus ketergantungan `github.com/gin-gonic/gin` secara menyeluruh dari codebase dan `go.mod`.
2. Menggunakan router standar Go 1.22+ `net/http.ServeMux` sebagai root multiplexer HTTP.
3. Menstandarkan semua HTTP middleware ke signature standar Go `func(http.Handler) http.Handler`.
4. Menggunakan `pkg/logger` (Logrus) untuk structured request logging dan panic recovery.

## Konsekuensi
- Handler dan transport sepenuhnya kompatibel dengan ekosistem standar Go `net/http`.
- ConnectRPC service handlers langsung di-mount ke `*http.ServeMux` tanpa wrapper.
- Kode streaming SSE dan WebSocket PTY terminal lebih bersih dan tidak terikat framework pihak ketiga.
- Dependency tree lebih ramping dan build binary lebih cepat.

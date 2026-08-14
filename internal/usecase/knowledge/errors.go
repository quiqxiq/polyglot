package knowledge

import "errors"

// ErrInvalidTitle menandai title kosong — field wajib untuk semua entry.
var ErrInvalidTitle = errors.New("knowledge: title is required")

// ErrEmptyContent menandai content kosong padahal dokumen diminta di-embed —
// raw-text AnythingLLM menolak textContent kosong (422), jadi dicegah lebih
// awal di usecase.
var ErrEmptyContent = errors.New("knowledge: content is required when embedding to AnythingLLM")

// ErrEmbedNotConfigured menandai embed diminta padahal AnythingLLM tidak
// dikonfigurasi (ANYTHINGLLM_API_KEY kosong) — manager adapter nil.
var ErrEmbedNotConfigured = errors.New("knowledge: AnythingLLM is not configured; cannot embed document")

// ErrEmbedSync membungkus kegagalan sinkronisasi ke AnythingLLM. Kalau error
// ini muncul, dokumen TETAP tersimpan di Postgres dengan embed_status
// "failed" — caller (handler) bisa mengecek errors.Is untuk menampilkan
// warning alih-alih menganggap operasi gagal total.
var ErrEmbedSync = errors.New("knowledge: AnythingLLM embed sync failed")

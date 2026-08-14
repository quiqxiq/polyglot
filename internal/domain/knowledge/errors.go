package knowledge

import "errors"

// ErrNotFound menandai entry knowledge tidak ditemukan. Dipakai repo
// (adapter/postgres) dan usecase (internal/usecase/knowledge) — pola sama
// dengan domain/device.ErrNotFound, sehingga usecase bisa memetakan
// not-found tanpa bergantung pada sentinel package adapter.
var ErrNotFound = errors.New("knowledge: entry not found")

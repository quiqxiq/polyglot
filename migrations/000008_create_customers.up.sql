CREATE TABLE customers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name           TEXT NOT NULL,
    id_number           TEXT,
    phone               TEXT NOT NULL,
    whatsapp            TEXT,
    email               TEXT,
    address             TEXT NOT NULL,
    location_lat        DOUBLE PRECISION,
    location_lng        DOUBLE PRECISION,
    customer_type       TEXT NOT NULL DEFAULT 'residential' CHECK (customer_type IN ('residential','business')),
    status              TEXT NOT NULL DEFAULT 'prospect' CHECK (status IN ('prospect','active','suspended','terminated')),
    referral_source     TEXT,
    notes               TEXT,
    registered_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    terminated_at       TIMESTAMPTZ
);

CREATE TABLE customer_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id     UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    doc_type        TEXT NOT NULL CHECK (doc_type IN ('ktp','kk','npwp','other')),
    file_url        TEXT NOT NULL,
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

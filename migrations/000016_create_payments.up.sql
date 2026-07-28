CREATE TABLE payments (
    id                              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id                     UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    payment_method                  TEXT NOT NULL CHECK (payment_method IN
                                        ('cash','bank_transfer','e_wallet','payment_gateway','retail_outlet')),
    payment_gateway_id              UUID REFERENCES payment_gateways(id) ON DELETE SET NULL,
    payment_gateway_transaction_id  TEXT,
    amount                          NUMERIC(14,2) NOT NULL,
    status                          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','confirmed','failed','refunded')),
    reference_number                TEXT,
    verified_by                     UUID REFERENCES users(id) ON DELETE SET NULL,
    paid_at                         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payment_allocations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id          UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    invoice_id          UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount_allocated    NUMERIC(14,2) NOT NULL,
    UNIQUE (payment_id, invoice_id)
);

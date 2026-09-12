-- 000030_document_number_sequences.up.sql
-- Sequence nomor dokumen PAY-/TRX- agar tidak pernah tabrakan antar transaksi
-- konkuren (F6-2).
CREATE SEQUENCE IF NOT EXISTS payments_no_seq;
CREATE SEQUENCE IF NOT EXISTS cash_transactions_no_seq;

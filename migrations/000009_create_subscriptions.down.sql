-- Lepas dulu FK yang ditutup di file .up ini sebelum drop tabelnya,
-- supaya urutan rollback benar (constraint menunjuk ke subscriptions).
ALTER TABLE hotspot_vouchers DROP CONSTRAINT fk_voucher_subscription;

DROP TABLE subscriptions;

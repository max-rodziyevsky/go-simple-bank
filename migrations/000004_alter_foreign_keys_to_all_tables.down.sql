ALTER TABLE "entries" DROP CONSTRAINT entries_account_id_fkey;

ALTER TABLE "transfers" DROP CONSTRAINT transfers_from_account_id_fkey;

ALTER TABLE "transfers" DROP CONSTRAINT transfers_to_account_id_fkey;

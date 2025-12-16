ALTER TABLE "accounts" ADD CONSTRAINT "balance_non_negative" CHECK (balance >= 0);
ALTER TABLE "transfers" ADD CONSTRAINT "amount_positive" CHECK (amount > 0);
ALTER TABLE "transfers" ADD CONSTRAINT "different_accounts" CHECK (from_account_id != to_account_id);

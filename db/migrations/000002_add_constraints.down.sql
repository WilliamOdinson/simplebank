ALTER TABLE "transfers" DROP CONSTRAINT IF EXISTS "different_accounts";
ALTER TABLE "transfers" DROP CONSTRAINT IF EXISTS "amount_positive";
ALTER TABLE "accounts" DROP CONSTRAINT IF EXISTS "balance_non_negative";

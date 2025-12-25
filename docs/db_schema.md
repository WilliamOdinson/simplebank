# Database Design

To make the basic functions run, Simple Bank supports services with four core functions: user management, account management, balance tracking, and money transfers.

**Users Table**
Stores user authentication and profile information. Each user has a unique `username` as the primary key, along with their hashed password, full name, and email. Tracks when the password was last changed and when the account was created.

**Accounts Table**
Stores customer account information. Each account has a unique ID, references an owner (linked to the users table), balance, currency, and creation timestamp. An index on `owner` allows fast lookups by account holder. A composite unique index on `(owner, currency)` ensures each user can only have one account per currency.

**Entries Table**
Logs every change in account balance. Each entry references an account via `account_id`, and records the change amount (positive for deposit, negative for withdrawal) with a timestamp. An index on `account_id` supports efficient retrieval of an account's transaction history.

**Transfers Table**
Captures money movement between two accounts. Contains references to both source (`from_account_id`) and destination (`to_account_id`) accounts, the positive transfer amount, and a timestamp. Indexed on `from_account_id`, `to_account_id`, and their combination for quick queries of transfers by account or account pair.

```mermaid
erDiagram
  USERS ||--o{ ACCOUNTS : "username -> owner"
  ACCOUNTS ||--o{ ENTRIES : "id -> account_id"
  ACCOUNTS ||--o{ TRANSFERS : "id -> from_account_id"
  ACCOUNTS ||--o{ TRANSFERS : "id -> to_account_id"

  USERS {
    VARCHAR username PK
    VARCHAR hashed_password
    VARCHAR full_name
    VARCHAR email UK
    TIMESTAMPTZ password_changed_at
    TIMESTAMPTZ created_at
  }

  ACCOUNTS {
    BIGSERIAL id PK
    VARCHAR owner FK
    BIGINT balance
    VARCHAR currency
    TIMESTAMPTZ created_at
  }

  ENTRIES {
    BIGSERIAL id PK
    BIGINT account_id FK
    BIGINT amount
    TIMESTAMPTZ created_at
  }

  TRANSFERS {
    BIGSERIAL id PK
    BIGINT from_account_id FK
    BIGINT to_account_id FK
    BIGINT amount
    TIMESTAMPTZ created_at
  }
```

Here's the [dbdiagram.io](https://dbdiagram.io/) script.

```sql
Table users as U {
  username varchar [pk]
  hashed_password varchar [not null]
  full_name varchar [not null]
  email varchar [unique, not null]
  password_changed_at timestamptz [not null, default: `0001-01-01 00:00:00Z`]
  created_at timestamptz [not null, default: `now()`]
}

Table accounts as A {
  id bigserial [pk]
  owner varchar [ref: > U.username, not null]
  balance bigint [not null]
  currency varchar [not null]
  created_at timestamptz [not null, default: `now()`]

  Indexes {
    owner
    (owner, currency) [unique]
  }
}

Table entries {
  id bigserial [pk]
  account_id bigint [ref: > A.id, not null]
  amount bigint [not null, note: 'can be negative or positive']
  created_at timestamptz [not null, default: `now()`]

  Indexes {
    account_id
  }
}

Table transfers {
  id bigserial [pk]
  from_account_id bigint [ref: > A.id, not null]
  to_account_id bigint [ref: > A.id, not null]
  amount bigint [not null, note: 'must be positive']
  created_at timestamptz [not null, default: `now()`]

  Indexes {
    from_account_id
    to_account_id
    (from_account_id, to_account_id)
  }
}
```

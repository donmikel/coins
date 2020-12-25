SELECT 'CREATE DATABASE coins'
WHERE NOT EXISTS(SELECT FROM pg_database WHERE datname = 'coins')
\gexec

CREATE TABLE IF NOT EXISTS accounts
(
    id       varchar(250) primary key,
    balance  decimal    NOT NULL,
    currency varchar(3) NOT NULL
);

CREATE TABLE IF NOT EXISTS payments
(
    id           bigserial primary key,
    from_account text     NOT NULL,
    to_account   text     NOT NULL,
    amount       numeric  NOT NULL,
    direction    smallint NOT NULL,
    dt           timestamp DEFAULT now()
);

CREATE OR REPLACE PROCEDURE send_payment_proc(from_account text, to_account text, amount decimal,
                                              direction smallint)
as
$$
DECLARE
    from_currency varchar(3);
    to_currency   varchar(3);
    u_count       int;
BEGIN
    SELECT a.currency
    INTO from_currency
    FROM accounts AS a
    WHERE a.id = from_account
        FOR UPDATE;

    SELECT a.currency
    INTO to_currency
    FROM accounts AS a
    WHERE a.id = to_currency
        FOR UPDATE;

    IF from_account <> to_currency THEN
        RETURN;
    end if;

    UPDATE accounts as a
    SET balance = balance - amount
    where a.id = from_account
      and a.balance - amount >= 0;
    get diagnostics u_count = row_count;
    IF u_count = 0 THEN
        ROLLBACK;
        RETURN;
    end if;

    UPDATE accounts as a
    SET balance = balance + amount
    where a.id = to_account;
    get diagnostics u_count = row_count;
    IF u_count = 0 THEN
        ROLLBACK;
        RETURN;
    end if;

    INSERT INTO payments (from_account, to_account, amount, direction)
    VALUES (from_account, to_account, amount, direction);

    COMMIT;
END;
$$
    LANGUAGE plpgsql;

INSERT INTO accounts
VALUES ('bob123', 100, 'USD'),
       ('alice456', 0.01, 'USD');
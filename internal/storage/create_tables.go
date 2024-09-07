package storage

import (
	"context"
)

func (dr *DBRepository) createUsersDB(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
		user_login varchar NOT NULL,
		user_password varchar NOT NULL,
		user_uuid varchar NOT NULL,
		user_balance numeric(8, 2) DEFAULT 0,
		user_withdrawn numeric(8, 2) DEFAULT 0, 
		CONSTRAINT users_pk PRIMARY KEY (user_uuid),
		CONSTRAINT login_unique UNIQUE (user_login)
		);`
	_, err := dr.db.Exec(ctx, query)
	return err
}

func (dr *DBRepository) createOrdersDB(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
		user_uuid varchar REFERENCES users ON DELETE CASCADE,
		order_id varchar NOT NULL,
		order_status varchar NOT NULL,
		order_uploaded_at timestamp with time zone NOT NULL,
		order_accrual numeric(8, 2),
		CONSTRAINT orders_pk PRIMARY KEY (order_id)
		);`
	_, err := dr.db.Exec(ctx, query)
	return err
}

func (dr *DBRepository) createWithdrawalsDB(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS withdrawals (
		order_id varchar NOT NULL,
		user_uuid varchar REFERENCES users ON DELETE CASCADE,
		processed_at timestamp with time zone NOT NULL,
		amount numeric(8, 2),
		CONSTRAINT withdrawals_pk PRIMARY KEY (order_id)
		);`
	_, err := dr.db.Exec(ctx, query)
	return err
}

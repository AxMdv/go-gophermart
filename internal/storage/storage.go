package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AxMdv/go-gophermart/internal/config"
	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBRepository struct {
	db *pgxpool.Pool
}

func NewRepository(ctx context.Context, config *config.Options) (*DBRepository, error) {
	pool, err := pgxpool.New(context.Background(), config.DataBaseURI)
	if err != nil {
		return &DBRepository{}, err
	}

	dbRepository := DBRepository{
		db: pool,
	}

	err = pool.Ping(ctx)
	if err != nil {
		return &dbRepository, err
	}

	err = dbRepository.createUsersDB(ctx)
	if err != nil {
		return &dbRepository, err
	}
	err = dbRepository.createOrdersDB(ctx)
	if err != nil {
		return &dbRepository, err
	}
	return &dbRepository, nil
}

func (dr *DBRepository) createUsersDB(ctx context.Context) error {
	var tableName string
	query1 := `
	SELECT table_name
	FROM information_schema.tables
	WHERE table_schema = 'public'
	AND table_name = 'users'
	LIMIT 1;`
	row := dr.db.QueryRow(ctx, query1)
	err := row.Scan(&tableName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			query2 := `
				CREATE TABLE users (
				user_login varchar NOT NULL,
				user_password varchar NOT NULL,
				user_uuid varchar NOT NULL,
				user_balance numeric(8, 2),
				user_withdrawn numeric(8, 2),
				CONSTRAINT users_pk PRIMARY KEY (user_uuid),
				CONSTRAINT login_unique UNIQUE (user_login)
				);`
			_, err := dr.db.Exec(ctx, query2)
			return err
		}
	}
	return err
}

func (dr *DBRepository) createOrdersDB(ctx context.Context) error {
	var tableName string
	query1 := `
	SELECT table_name
	FROM information_schema.tables
	WHERE table_schema = 'public'
	AND table_name = 'orders'
	LIMIT 1;`
	row := dr.db.QueryRow(ctx, query1)
	err := row.Scan(&tableName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			query2 := `
				CREATE TABLE orders (
				user_uuid varchar REFERENCES users ON DELETE CASCADE,
				order_id varchar NOT NULL,
				order_status varchar NOT NULL,
				order_uploaded_at timestamp with time zone NOT NULL,
				order_accrual numeric(8, 2),
				CONSTRAINT orders_pk PRIMARY KEY (order_id)
				);`
			_, err := dr.db.Exec(ctx, query2)
			return err
		}
	}
	return err
}
func (dr *DBRepository) RegisterUser(ctx context.Context, user *model.User) error {
	query := `
	INSERT INTO users (user_login, user_password, user_uuid)
	VALUES ($1, $2, $3);`
	_, err := dr.db.Exec(ctx, query, user.Login, user.Password, user.UUID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return ErrLoginDuplicate
			}
		}
		return err
	}
	return nil

}

func (dr *DBRepository) GetUserAuthData(ctx context.Context, reqUser *model.User) (dbUser *model.User, err error) {
	dbUser = &model.User{}
	query := `
	SELECT user_login, user_password, user_uuid 
	FROM users WHERE user_login = $1`
	row := dr.db.QueryRow(ctx, query, reqUser.Login)
	err = row.Scan(&dbUser.Login, &dbUser.Password, &dbUser.UUID)
	if err != nil {
		// may be it s not necessary
		if errors.Is(err, pgx.ErrNoRows) {
			return dbUser, ErrNoAuthData
		}
	}
	return dbUser, err
}

func (dr *DBRepository) GetOrderByID(ctx context.Context, order *model.Order) (userID string, err error) {
	userID = ""
	query := `
	SELECT user_uuid
	FROM orders WHERE order_id = $1;`
	row := dr.db.QueryRow(ctx, query, order.ID)
	err = row.Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userID, ErrNoOrder
		}
	}
	return userID, err
}

func (dr *DBRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	query := `
	INSERT INTO orders (order_id, user_uuid, order_uploaded_at, order_status)
	VALUES ($1, $2, $3, $4);`
	_, err := dr.db.Exec(ctx, query, order.ID, order.UserUUID, order.UploadedAt, order.Status)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				fmt.Println(err, pgErr.Code, pgErr.Detail)
				return ErrOrderDuplicate
			}
		}
		return err
	}
	return nil
}

func (dr *DBRepository) GetOrdersByUserID(ctx context.Context, userID string) ([]model.Order, error) {
	orders := make([]model.Order, 0, 10)
	query := `SELECT order_id, order_status, order_accrual, order_uploaded_at
	FROM orders WHERE user_uuid = $1
	ORDER BY order_uploaded_at DESC;`
	rows, _ := dr.db.Query(ctx, query, userID)
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		var accrual sql.NullFloat64
		err := rows.Scan(&order.ID, &order.Status, &accrual, &order.UploadedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNoOrders
			}
			return nil, err
		}
		if accrual.Valid {
			order.Accrual = accrual.Float64
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// func (dr *DBRepository) (ctx context.Context, user *model.User) error {}
// func (dr *DBRepository) (ctx context.Context, user *model.User) error {}
// func (dr *DBRepository) (ctx context.Context, user *model.User) error {}

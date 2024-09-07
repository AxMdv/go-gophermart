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

func NewRepository(ctx context.Context, config *config.Config) (*DBRepository, error) {
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
	err = dbRepository.createWithdrawalsDB(ctx)
	if err != nil {
		return &dbRepository, err
	}
	return &dbRepository, nil
}

func (dr *DBRepository) RegisterUser(ctx context.Context, user *model.User) error {
	query := `
	INSERT INTO users (user_login, user_password, user_uuid)
	VALUES ($1, $2, $3);`
	_, err := dr.db.Exec(ctx, query, user.Login, user.Password, user.UUID)
	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return ErrLoginDuplicate
	}
	return err
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

func (dr *DBRepository) GetUserBalance(ctx context.Context, userID string) (*model.Balance, error) {
	query := `
	SELECT user_balance, user_withdrawn
	FROM users
	WHERE user_uuid = $1;`
	row := dr.db.QueryRow(ctx, query, userID)

	balance := model.Balance{}
	err := row.Scan(&balance.Current, &balance.Withdrawn)
	return &balance, err
}

func (dr *DBRepository) GetWithdrawalsByUserID(ctx context.Context, userID string) ([]model.Withdrawal, error) {
	query := `
	SELECT order_id, processed_at, amount
	FROM withdrawals
	WHERE user_uuid = $1
	ORDER BY processed_at DESC;`
	rows, _ := dr.db.Query(ctx, query, userID)
	defer rows.Close()

	withdrawals := make([]model.Withdrawal, 0, 5)
	for rows.Next() {
		var wdwl model.Withdrawal
		err := rows.Scan(&wdwl.OrderID, &wdwl.ProcessedAt, &wdwl.Amount)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNoWithdrawalsData
			}
			return nil, err
		}
		withdrawals = append(withdrawals, wdwl)

	}
	return withdrawals, nil

}

func (dr *DBRepository) CreateWithdraw(ctx context.Context, balance *model.Balance, withdrawal *model.Withdrawal) error {
	query1 := `
	UPDATE users
	SET user_balance = $1, user_withdrawn = $2
	WHERE user_uuid = $3;`

	query2 := `
	INSERT INTO withdrawals (order_id, user_uuid, processed_at, amount) 
	VALUES ($1, $2, $3, $4);`

	tx, err := dr.db.Begin(ctx)

	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, query1, balance.Current, balance.Withdrawn, balance.UserUUID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, query2, withdrawal.OrderID, withdrawal.UserUUID, withdrawal.ProcessedAt, withdrawal.Amount)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
	// dr.db.Exec(ctx, query1, balance.Current, balance.Withdrawn, balance.UserUUID)

}

func (dr *DBRepository) UpdateUserBalance(ctx context.Context, order *model.Order) error {

	query1 := `
	UPDATE users
	SET user_balance = user_balance + $1
	WHERE user_uuid = $2;`
	_, err := dr.db.Exec(ctx, query1, order.Accrual, order.UserUUID)
	if err != nil {
		return err
	}
	return nil
}

func (dr *DBRepository) UpdateOrder(ctx context.Context, order *model.Order) error {
	query1 := `
	UPDATE orders
	SET order_status = $1, order_accrual = $2
	WHERE order_id = $3;`
	_, err := dr.db.Exec(ctx, query1, order.Status, order.Accrual, order.ID)
	if err != nil {
		return err
	}
	return nil
}

// func (dr *DBRepository) (ctx context.Context, user *model.User) error {}

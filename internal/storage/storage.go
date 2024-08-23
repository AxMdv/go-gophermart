package storage

import (
	"context"
	"errors"

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
				login varchar NOT NULL,
				password varchar NOT NULL,
				uuid varchar NOT NULL,
				CONSTRAINT users_pk PRIMARY KEY (login),
				CONSTRAINT uuid_unique UNIQUE (uuid)
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
				users_uuid varchar NOT NULL,
				order_id varchar NOT NULL,
				uuid varchar NOT NULL,
				CONSTRAINT users_pk PRIMARY KEY (login),
				CONSTRAINT uuid_unique UNIQUE (uuid)
				);`
			_, err := dr.db.Exec(ctx, query2)
			return err
		}
	}
	return err
}
func (dr *DBRepository) RegisterUser(ctx context.Context, user *model.User) error {
	query := `
	INSERT INTO users (login, password, uuid)
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
	SELECT login, password, uuid 
	FROM users WHERE login = $1`
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

// func (dr *DBRepository) (ctx context.Context, user *model.User) error {}
// func (dr *DBRepository) (ctx context.Context, user *model.User) error {}
// func (dr *DBRepository) (ctx context.Context, user *model.User) error {}

package user

import (
	"context"
	"errors"
	"fmt"
	"AuthService/pkg/core/token"
	"AuthService/pkg/mux/middleware/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

type ResponseDTO struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

func (s *Service) Profile(ctx context.Context) (response ResponseDTO, err error) {
	auth, ok := jwt.FromContext(ctx).(*token.Payload)
	if !ok {
		return ResponseDTO{}, errors.New("bad request")
	}

	return ResponseDTO{
		Id:     auth.Id,
		Name:   auth.Username,
		Avatar: "https://i.pravatar.cc/50",
	}, nil
}

func (s *Service) FindUserByID(id int64, pool *pgxpool.Pool) (response ResponseDTO, err error) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		err = errors.New("server error")
		return
	}
	defer conn.Release()
	var (
		username string
		isAdmin  bool
		removed  bool
	)
	fmt.Println(id)
	err = conn.QueryRow(context.Background(), `select username, admin, removed from users where id = $1;`, id).Scan(&username, &isAdmin, &removed)
	if err != nil {
		err = errors.New("no such user")
		return
	}
	if isAdmin {
		err = errors.New("you can't delete admin")
		return
	}
	if removed {
		err = errors.New("this user already deleted")
		return
	}

	return ResponseDTO{
		Id:     id,
		Name:   username,
		Avatar: "https://i.pravatar.cc/50",
	}, nil
}

func (s *Service) DelUserByID(id int64, pool *pgxpool.Pool) (err error) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		err = errors.New("server error")
		return
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), `UPDATE users SET removed = True WHERE id = $1;`, id)
	if err != nil {
		err = errors.New("server error")
		return
	}
	return
}

func (s *Service) RegisterUser(newUser token.RequestDTO, pool *pgxpool.Pool) (err error) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		err = errors.New("server error")
		return
	}
	defer conn.Release()
	password, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		err = errors.New("server error")
		return
	}
	_, err = conn.Exec(context.Background(), `insert into users(username, password) Values ($1, $2);`, newUser.Username, password)
	if err != nil {
		err = errors.New("server error")
		return
	}
	return
}

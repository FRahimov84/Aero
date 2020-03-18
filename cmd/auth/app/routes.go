package app

import (
	"context"
	"errors"
	"github.com/FRahimov84/AuthService/pkg/core/token"
	"github.com/FRahimov84/AuthService/pkg/mux/middleware/authenticated"
	"github.com/FRahimov84/AuthService/pkg/mux/middleware/authorized"
	"github.com/FRahimov84/AuthService/pkg/mux/middleware/jwt"
	"github.com/FRahimov84/AuthService/pkg/mux/middleware/logger"
	"reflect"
)

func (s *Server) InitRoutes() {

	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		panic(errors.New("can't create database"))
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), `
CREATE TABLE if not exists users (
   id BIGSERIAL PRIMARY KEY,
   username TEXT NOT NULL unique,
   password TEXT NOT NULL,
   admin BOOLEAN DEFAULT FALSE,
   removed BOOLEAN DEFAULT FALSE
);
`)
	if err != nil {
		panic(errors.New("can't create database"))
	}

	s.router.GET(
		"/",
		s.handleIndex(),
		logger.Logger("Index"),
	)
	//s.router.GET(
	//	"/login",
	//	s.handleAdminLogin(),
	//	logger.Logger("login Admin Panel"),
	//)

	s.router.POST(
		"/api/tokens",
		s.handleCreateToken(),
		//authenticated.Authenticated(jwt.IsContextNonEmpty),
		//tokens.HandleCreateToken(s),
		logger.Logger("TOKEN"),
	)
	//s.router.POST(
	//	"/api/admin",
	//	s.handAdmin(),
	//	authorized.Authorized([]string{"Admin"}, jwt.FromContext),
	//	authenticated.Authenticated(jwt.IsContextNonEmpty),
	//	logger.Logger("admin"),
	//)
	// /api/users/me
	// golang нельзя reflect.TypeOf(token.Payload)
	// golang нельзя reflect.TypeOf((*token.Payload)(nil))
	s.router.GET(
		"/api/users/me",
		s.handleProfile(),
		authenticated.Authenticated(jwt.IsContextNonEmpty),
		jwt.JWT(reflect.TypeOf((*token.Payload)(nil)).Elem(), s.secret),
		logger.Logger("USERS/me"),
	)

	s.router.DELETE(
		"/api/users/{id}",
		s.handleDeleteProfile(),
		authenticated.Authenticated(jwt.IsContextNonEmpty),
		authorized.Authorized([]string{"Admin"}, jwt.FromContext),
		jwt.JWT(reflect.TypeOf((*token.Payload)(nil)).Elem(), s.secret),
		logger.Logger("USER Delete"),
	)
	s.router.POST(
		"/api/users/new",
		s.handleRegister(),
		authenticated.Authenticated(jwt.IsContextNonEmpty),
		authorized.Authorized([]string{"Admin"}, jwt.FromContext),
		jwt.JWT(reflect.TypeOf((*token.Payload)(nil)).Elem(), s.secret),
		logger.Logger("USER Register"),
	)
}

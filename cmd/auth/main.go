package main

import (
	"auth-service/cmd/auth/app"
	"auth-service/pkg/core/token"
	"auth-service/pkg/core/user"
	"context"
	"flag"
	"fmt"
	"github.com/FRahimov84/Mux/pkg/mux"
	"github.com/FRahimov84/di/pkg/di"
	"github.com/FRahimov84/myJwt/pkg/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"net"
	"net/http"
	"os"
	"path/filepath"
)



var (
	host = flag.String("host", "", "Server host")
	port = flag.String("port", "", "Server port")
	dsn  = flag.String("dsn", "", "Postgres DSN")
)

const (
	envHost = "HOST"
	envPort = "PORT"
	envDSN  = "DATABASE_URL"
)

type DSN string

func main() {
	flag.Parse()
	serverHost := checkENV(envHost, *host)
	serverPort := checkENV(envPort, *port)
	serverDsn := checkENV(envDSN, *dsn)
	addr := net.JoinHostPort(serverHost, serverPort)
	secret := jwt.Secret("secret")
	start(addr, serverDsn, secret)
}

func checkENV(env string, loc string) string {
	str, ok := os.LookupEnv(env)
	if !ok {
		return loc
	}
	return str
}

func start(addr string, dsn string, secret jwt.Secret) {

	container := di.NewContainer()

	container.Provide(
		app.NewServer,
		mux.NewExactMux,
		func() app.TempPath { return app.TempPath(filepath.Join("web", "templates")) },
		func() app.AssetsPath { return app.AssetsPath(filepath.Join("web", "assets")) },
		func() jwt.Secret { return secret },
		func() DSN { return DSN(dsn) },
		func(dsn DSN) *pgxpool.Pool {
			pool, err := pgxpool.Connect(context.Background(), string(dsn))
			if err != nil {
				panic(fmt.Errorf("can't create pool: %w", err))
			}
			return pool
		},
		token.NewService,
		user.NewService,
	)

	container.Start()
	// IoC - inversion of control (программа определяет, куда вы можете встроиться)
	// StartListener, StopListener
	// см. Errors.As
	var appServer *app.Server
	container.Component(&appServer)
	//http.Handle("/assets/",http.StripPrefix("/assets/",http.FileServer(http.Dir("./assets/"))))
	//http.HandleFunc("/", indexHandler)
	panic(http.ListenAndServe(addr, appServer))
}

//func indexHandler(writer http.ResponseWriter, request *http.Request) {
//	tpl, err := template.ParseFiles(
//		filepath.Join("web/templates", "index.gohtml"),
//		filepath.Join("web/templates", "header.gohtml"),
//		filepath.Join("web/templates", "footer.gohtml"))
//	if err != nil {
//		log.Printf("can't ParseFiles in func handleIndex err: %v", err)
//		panic(err)
//	}
//	tpl.ExecuteTemplate(writer, "index", nil)
//}

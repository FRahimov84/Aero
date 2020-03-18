package app

import (
	"auth-service/pkg/core/token"
	"auth-service/pkg/core/user"
	"fmt"
	"github.com/FRahimov84/Mux/pkg/mux"
	"github.com/FRahimov84/myJwt/pkg/jwt"
	"github.com/FRahimov84/rest/pkg/rest"
	"github.com/jackc/pgx/v4/pgxpool"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

type (
	TempPath   string
	AssetsPath string
)
type Server struct {
	router         *mux.ExactMux
	pool           *pgxpool.Pool
	secret         jwt.Secret
	tokenSvc       *token.Service
	userSvc        *user.Service
	templatesPath  TempPath
	assetsPath     AssetsPath
}

func NewServer(router *mux.ExactMux, pool *pgxpool.Pool, secret jwt.Secret, tokenSvc *token.Service, userSvc *user.Service, templatesPath TempPath, assetsPath AssetsPath) *Server {
	return &Server{router: router, pool: pool, secret: secret, tokenSvc: tokenSvc, userSvc: userSvc, templatesPath: templatesPath, assetsPath: assetsPath}
}

func (s *Server) Start() {
	s.InitRoutes()
}

func (s *Server) Stop() {
	// TODO: make server stop
}

type ErrorDTO struct {
	Errors []string `json:"errors"`
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s *Server) handleCreateToken() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var body token.RequestDTO
		//all, _ := ioutil.ReadAll(request.Body)
		//fmt.Println(string(all))

		err := rest.ReadJSONBody(request, &body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			err := rest.WriteJSONBody(writer, &ErrorDTO{
				[]string{"err.json_invalid"},
			})
			log.Print(err)
			return
		}

		response, err := s.tokenSvc.Generate(request.Context(), &body, s.pool)

		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			err2 := rest.WriteJSONBody(writer, &ErrorDTO{
				[]string{"err.password_mismatch", err.Error()},
			})
			if err2 != nil {
				log.Print(err2)
			}
			return
		}

		err = rest.WriteJSONBody(writer, &response)
		if err != nil {
			log.Print(err)
		}
	}
}

func (s *Server) handleDeleteProfile() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		id, err := strconv.Atoi(request.URL.Path[11:])
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		byID, err := s.userSvc.FindUserByID(int64(id), s.pool)
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
		profile, err := s.userSvc.Profile(request.Context())
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if byID.Name == profile.Name{
			writer.Write([]byte("you can't delete yourself"))
			return
		}
		err = s.userSvc.DelUserByID(int64(id), s.pool)
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
		writer.Write([]byte("done!"))
	}
}

func (s *Server) handleIndex() http.HandlerFunc {
		// executes in one goroutine
		var (
		tpl *template.Template
		err error
	)
		tpl, err = template.ParseFiles(
			filepath.Join("web/templates", "index.gohtml"),
			//filepath.Join("web/templates", "header.gohtml"),
			//filepath.Join("web/templates", "footer.gohtml"),
			)
		if err != nil {
		panic(err)
	}
		return func(writer http.ResponseWriter, request *http.Request) {
		// executes in many goroutines
		// TODO: fetch data from multiple upstream services
		err := tpl.Execute(writer, struct {Title string}{Title: "Auth Service",})
		if err != nil {
		log.Printf("error while executing template %s %v", tpl.Name(), err)
	}
	}

}

func (s *Server) handleRegister() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		get := request.Header.Get("Content-Type")
		fmt.Println(get)
		if get != "application/json" {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var newUser token.RequestDTO

		err := rest.ReadJSONBody(request, &newUser)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err = s.userSvc.RegisterUser(newUser, s.pool)
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
		writer.Write([]byte("done!"))

	}
}

func (s *Server) handleProfile() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		response, err := s.userSvc.Profile(request.Context())
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			err := rest.WriteJSONBody(writer, &ErrorDTO{
				[]string{"err.bad_request"},
			})
			log.Print(err)
			return
		}
		err = rest.WriteJSONBody(writer, &response)
		if err != nil {
			log.Print(err)
		}

	}
}
//
//func (s *Server) handleAdminLogin() http.HandlerFunc {
//	var (
//		tpl *template.Template
//		err error
//	)
//	tpl, err = template.ParseFiles(
//		filepath.Join("web/templates", "AdminPanel.gohtml"),
//		//filepath.Join("web/templates", "header.gohtml"),
//		//filepath.Join("web/templates", "footer.gohtml"),
//	)
//	if err != nil {
//		panic(err)
//	}
//	return func(writer http.ResponseWriter, request *http.Request) {
//		err := tpl.Execute(writer, struct {Title string}{Title: "Admin Panel",})
//		if err != nil {
//			log.Printf("error while executing template %s %v", tpl.Name(), err)
//		}
//	}
//}

//func (s *Server) handAdmin() http.HandlerFunc {
//
//}

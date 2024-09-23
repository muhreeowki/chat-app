package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/cors"
)

type JSONServer struct {
	store      Storage
	listenAddr string
}

func NewJSONServer(listenAddr string, store Storage) *JSONServer {
	return &JSONServer{
		store:      store,
		listenAddr: listenAddr,
	}
}

func (s *JSONServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /messages", withJWTAuth(makeHttpHandler(s.HandleGetMessages)))
	mux.HandleFunc("GET /users", withJWTAuth(makeHttpHandler(s.HandleGetUsers)))
	mux.HandleFunc("POST /signup", makeHttpHandler(s.HandleSignUp))
	mux.HandleFunc("POST /login", makeHttpHandler(s.HandleLogin))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3000/login"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	handler := c.Handler(mux)

	fmt.Printf("Mchat JSON server is live on: %s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, handler)
}

func (s *JSONServer) HandleGetMessages(w http.ResponseWriter, r *http.Request) *JSONServerError {
	messages, err := s.store.GetMessages()
	if err != nil {
		return &JSONServerError{
			code:  500,
			error: err.Error(),
		}
	}
	WriteJSON(w, http.StatusOK, messages)
	return nil
}

func (s *JSONServer) HandleSignUp(w http.ResponseWriter, r *http.Request) *JSONServerError {
	reqData, err := UnmarshalUserJSON(r)
	if err != nil {
		fmt.Println(err)
		return &JSONServerError{
			code:  http.StatusBadRequest,
			error: "invalid request data",
		}
	}
	hashedPass, err := HashPassword(reqData.Password)
	if err != nil {
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	reqData.Password = hashedPass
	// Create User
	if err := s.store.CreateUser(reqData); err != nil {
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	usr := &UserJSONResponse{
		Id:       reqData.Id,
		Username: reqData.Username,
	}
	usr.Token, err = createJWT(reqData)
	if err != nil {
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	// Return the user object and the jwt token
	WriteJSON(w, http.StatusCreated, usr)
	return nil
}

func (s *JSONServer) HandleLogin(w http.ResponseWriter, r *http.Request) *JSONServerError {
	reqData, err := UnmarshalUserJSON(r)
	if err != nil {
		return &JSONServerError{
			code:  500,
			error: err.Error(),
		}
	}
	// Check if user exists
	dbUsr, err := s.store.GetUser(reqData.Username)
	if err != nil {
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}
	// Validate Password
	valid := VerifyPassword(reqData.Password, dbUsr.Password)
	if !valid {
		return &JSONServerError{
			code:  http.StatusUnauthorized,
			error: "invalid credentials",
		}
	}

	usr := &UserJSONResponse{
		Username: dbUsr.Username,
		Id:       dbUsr.Id,
	}
	usr.Token, err = createJWT(dbUsr)
	if err != nil {
		return &JSONServerError{
			code:  http.StatusInternalServerError,
			error: err.Error(),
		}
	}

	WriteJSON(w, http.StatusOK, usr)
	return nil
}

func (s *JSONServer) HandleGetUsers(w http.ResponseWriter, r *http.Request) *JSONServerError {
	usrs, err := s.store.GetUsers()
	if err != nil {
		return &JSONServerError{
			code:  500,
			error: err.Error(),
		}
	}
	WriteJSON(w, http.StatusOK, usrs)
	return nil
}

type JSONServerFunc func(w http.ResponseWriter, r *http.Request) *JSONServerError

type CustomClaims struct {
	Username string
	Password string
	jwt.RegisteredClaims
}

type JSONServerError struct {
	error string
	code  int
}

func (e *JSONServerError) Error() string {
	return e.error
}

func makeHttpHandler(serverFunc JSONServerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := serverFunc(w, r); err != nil {
			WriteJSON(w, err.code, err.Error())
			return
		}
	}
}

func withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("checking JWT Token")

		tokenString := r.Header.Get("Authorization")
		_, err := validateJWT(tokenString)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, JSONServerError{
				code:  http.StatusUnauthorized,
				error: "invalid token",
			})
			return
		}

		handlerFunc(w, r)
	}
}

func createJWT(usr *User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	// Create JWT
	claims := &CustomClaims{
		Username: usr.Username,
		Password: usr.Password,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        strconv.Itoa(usr.Id),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unnexpected signing method: %s", t.Method.Alg())
		}

		return []byte(secret), nil
	})
}

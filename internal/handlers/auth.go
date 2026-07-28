package handlers

import (
	"errors"
	"net/http"
	"weblog/internal/auth"
	"weblog/internal/repository"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

type signReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthHandler struct {
	UserRepo *repository.UserRepository
	JwtKey   string
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtKey string) (auth *AuthHandler) {
	auth = &AuthHandler{
		UserRepo: userRepo,
		JwtKey:   jwtKey,
	}
	return
}

func (ah *AuthHandler) SignUp(c echo.Context) error {
	var signUpReq signReq
	err := c.Bind(&signUpReq)
	if err != nil {
		return err
	}
	if signUpReq.Password == "" || signUpReq.Username == "" {
		return errors.New("invalid username or password")
	}

	if ah.UserRepo.IsTakenThisUsername(c.Request().Context(), signUpReq.Username) {
		return errors.New("taken this username! use another one")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(signUpReq.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	signUpReq.Password = string(hashedPass)
	user, err := ah.UserRepo.Add(c.Request().Context(), signUpReq.Username, signUpReq.Password)
	if err != nil {
		return err
	}

	claims := auth.CreateClaims(user.ID)
	token, err := auth.CreateToken(claims, ah.JwtKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"user":  user.ID,
		"token": token,
	})}

func (ah *AuthHandler) Login(c  echo.Context) error {
	var login signReq
	err := c.Bind(&login)
	if err != nil {
		return err
	}
	if login.Password == "" || login.Username == "" {
		return errors.New("invalid username or password")
	}

	user, err := ah.UserRepo.FindByUserName(c.Request().Context(), login.Username)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		return err
	}

	claims := auth.CreateClaims(user.ID)
	token, err := auth.CreateToken(claims, ah.JwtKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"user":  user.ID,
		"token": token,
	})

}

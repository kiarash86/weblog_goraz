package handlers

import (
	"net/http"
	"strconv"
	"weblog/internal/repository"

	"github.com/labstack/echo/v5"
)

type BoardShareHandler struct {
	BoardRepo      *repository.BoardRepository
	BoardShareRepo *repository.BoardShareRepository
	UserRepo       *repository.UserRepository
}

type CreateBoardSharedReq struct {
	Usernames []string `json:"usernames"`
}

func NewBoardShareHandler(br *repository.BoardRepository, bsr *repository.BoardShareRepository, ru *repository.UserRepository) *BoardShareHandler {
	boardShareHandler := &BoardShareHandler{
		BoardRepo:      br,
		BoardShareRepo: bsr,
		UserRepo:       ru,
	}

	return boardShareHandler
}

func (bsh *BoardShareHandler) Add(c *echo.Context) error {
	userID := c.Get("user_id").(int)
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	var reqCreate CreateBoardSharedReq
	err = c.Bind(&reqCreate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if len(reqCreate.Usernames) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "empty usernames")

	}
	board, err := bsh.BoardRepo.FindByID(c.Request().Context(), boardID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "couldnt get board")

	}
	if board == nil {
		return echo.NewHTTPError(http.StatusNotFound, "board not found")
	}
	if board.AuthorID != userID {

		return echo.NewHTTPError(http.StatusForbidden, "you dont  own this board")

	}

	usersFound := make([]string, 0)
	usersNotFound := make([]string, 0)
	for _, u := range reqCreate.Usernames {
		user, err := bsh.UserRepo.FindByUserName(c.Request().Context(), u)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong")
		}
		if user == nil {
			usersNotFound = append(usersNotFound, u)
			continue
		}

		usersFound = append(usersFound, u)
		_, err = bsh.BoardShareRepo.Add(c.Request().Context(), user.ID, boardID)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "couldnt add this "+u)
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"founded seccusfully": usersFound,
		"not founded":         usersNotFound,
	})

}

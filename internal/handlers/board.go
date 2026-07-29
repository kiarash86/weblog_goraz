package handlers

import (
	"net/http"
	"strconv"
	"weblog/internal/repository"

	"github.com/labstack/echo/v5"
)

type BoardHandler struct {
	BoardRepo      *repository.BoardRepository
	BoardShareRepo *repository.BoardShareRepository
}

type CreateBoardReq struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	IsPrivate bool   `json:"is_private"`
	ImgPath   string `json:"img_path"`
}

func NewBoardHandler(br *repository.BoardRepository, bsr *repository.BoardShareRepository) *BoardHandler {
	boardHandler := &BoardHandler{
		BoardRepo:      br,
		BoardShareRepo: bsr,
	}

	return boardHandler
}

func (bh *BoardHandler) Create(c *echo.Context) error {
	authorID := c.Get("user_id").(int)

	var reqCreate CreateBoardReq
	err := c.Bind(&reqCreate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if reqCreate.Title == "" || reqCreate.Content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty title and content")

	}

	b, err := bh.BoardRepo.Add(c.Request().Context(), authorID, reqCreate.Title, reqCreate.Content, reqCreate.IsPrivate, reqCreate.ImgPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "couldnt add this board")

	}
	return c.JSON(http.StatusCreated, b)
}

func (bh *BoardHandler) Feed(c *echo.Context) error {
	userID := c.Get("user_id").(int)

	list, err := bh.BoardRepo.ListFeed(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "couldnt get feed")
	}
	return c.JSON(http.StatusOK, list)
}

func (bh *BoardHandler) GetByID(c *echo.Context) error {
	userID := c.Get("user_id").(int)
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "couldnt get id")

	}

	board, err := bh.BoardRepo.FindByID(c.Request().Context(), boardID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong at getting board")

	}
	if board == nil {
		return echo.NewHTTPError(http.StatusNotFound, "board not found")
	}
	if board.ISPrivate {
		isOwner, err := bh.BoardRepo.IsOwner(c.Request().Context(), boardID, userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong with checking owner")

		}
		if isOwner {
			board, err := bh.BoardRepo.FindByID(c.Request().Context(), boardID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "couldnt get board")

			}
			return c.JSON(http.StatusOK, board)

		}
		hasAccess, err := bh.BoardShareRepo.HasAccess(c.Request().Context(), userID, boardID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong with ckecking access")

		}

		if hasAccess {
			board, err := bh.BoardRepo.FindByID(c.Request().Context(), boardID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "couldnt get board")

			}

			return c.JSON(http.StatusOK, board)
		} else {

			return echo.NewHTTPError(http.StatusForbidden, "doesnt have access to this post")
		}
	} else {
		return c.JSON(http.StatusOK, board)

	}
}

func (bh *BoardHandler) DeleteByID(c *echo.Context) error {
	userID := c.Get("user_id").(int)
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "couldnt get id")

	}

	isOwner, err := bh.BoardRepo.IsOwner(c.Request().Context(), boardID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())

	}
	if isOwner {

		err = bh.BoardRepo.DeleteByID(c.Request().Context(), boardID, userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

	} else {
		return echo.NewHTTPError(http.StatusForbidden, "u are not the owner")

	}

	return c.NoContent(http.StatusNoContent)

}

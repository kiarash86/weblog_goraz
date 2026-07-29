package handlers

import (
	"net/http"
	"strconv"
	"weblog/internal/repository"

	"github.com/labstack/echo/v5"
)

type CommentHandler struct {
	BoardRepo      *repository.BoardRepository
	BoardShareRepo *repository.BoardShareRepository
	CommentRepo    *repository.CommentRepository
}

type CreateCommentdReq struct {
	Content string `json:"content"`
}

func NewCommentHandler(br *repository.BoardRepository, bsr *repository.BoardShareRepository, rc *repository.CommentRepository) *CommentHandler {
	commentHandler := &CommentHandler{
		BoardRepo:      br,
		BoardShareRepo: bsr,
		CommentRepo:    rc,
	}

	return commentHandler
}

func (ch *CommentHandler) Create(c *echo.Context) error {
	userID := c.Get("user_id").(int)
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	var reqCreate CreateCommentdReq
	err = c.Bind(&reqCreate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if reqCreate.Content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty title and content")

	}
	board, err := ch.BoardRepo.FindByID(c.Request().Context() , boardID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "couldnt get board")
		
	}
	if board == nil {
	return echo.NewHTTPError(http.StatusNotFound, "board not found")
}
	if board.ISPrivate && board.AuthorID != userID {
		hasAccess, err := ch.BoardShareRepo.HasAccess(c.Request().Context(), userID, boardID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong")
		}
		if !hasAccess {
			return echo.NewHTTPError(http.StatusForbidden, "you dont have access to this board")
		}
	}



	b, err := ch.CommentRepo.Add(c.Request().Context(), userID, boardID, reqCreate.Content)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "couldnt add this board")

	}
	return c.JSON(http.StatusCreated, b)
}

func (ch *CommentHandler) ListCommentsOfBoard(c *echo.Context) error {
	userID := c.Get("user_id").(int)
boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}


		board, err := ch.BoardRepo.FindByID(c.Request().Context(), boardID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "couldnt get board")
	}
	if board == nil {
	return echo.NewHTTPError(http.StatusNotFound, "board not found")
}

	if board.ISPrivate && board.AuthorID != userID {
		hasAccess, err := ch.BoardShareRepo.HasAccess(c.Request().Context(), userID, boardID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong")
		}
		if !hasAccess {
			return echo.NewHTTPError(http.StatusForbidden, "you dont have access to this board")
		}
	}



	list, err := ch.CommentRepo.ListCommentsOfBoard(c.Request().Context(), boardID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "couldnt get feed")
	}
	return c.JSON(http.StatusOK, list)
}

func (ch *CommentHandler) DeleteByID(c *echo.Context) error {
	userID := c.Get("user_id").(int)
	commentID, err := strconv.Atoi(c.Param("commentId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "couldnt get id")

	}
	comment, err := ch.CommentRepo.FindByID(c.Request().Context(), commentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "couldnt get id")
	}
	if comment == nil {
	return echo.NewHTTPError(http.StatusNotFound, "comment not found")
}
	if comment.AuthorID == userID {
		err = ch.CommentRepo.DeleteByID(c.Request().Context(), commentID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())

		}
		return  c.NoContent(http.StatusNoContent)
	} else {
		return echo.NewHTTPError(http.StatusForbidden, "u are not the owner of this comment")

	}
}

package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
)
var imageTypeSupport = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}
type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (uh *UploadHandler) Upload(c *echo.Context) error {
	img, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "there is no image")
	}

	if img.Size > 5 * 1024 * 1024 {
		return echo.NewHTTPError(http.StatusBadRequest, "more than 5mg cant accept")
	}
	passwand := strings.ToLower(filepath.Ext(img.Filename))
	if !imageTypeSupport[passwand] {
		return echo.NewHTTPError(http.StatusBadRequest, "cant allow this type of images")
	}

	imageSource, err := img.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "there is something wrong with this img")
	}
	defer imageSource.Close()

	err = os.MkdirAll("uploads", 0755)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "no folder for upload")
	}


rnd := make([]byte, 16)
	_, err = rand.Read(rnd)
	if  err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "no unique name for this image")
	}
	filename := fmt.Sprintf("%s%s", hex.EncodeToString(rnd), passwand)


	path := filepath.Join("uploads", filename)
	file, err := os.Create(path)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cant add this image to folder uploads")
	}
	defer file.Close()
	 _, err = io.Copy(file, imageSource)	
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cant save this image")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"path": "/" + "uploads" + "/" + filename,
	})
}



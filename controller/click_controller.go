package controller

import (
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"net/http"
	"study-service/click/service"
	"study-service/common/errors"
	requestDto "study-service/dto/request"
)

type ClickController struct {
}

func (controller ClickController) Init(g *echo.Group) {
	g.POST("", controller.Create)
}

func (ClickController) Create(ctx echo.Context) error {
	var clickCreate = requestDto.ClickCreate{}

	if err := ctx.Bind(&clickCreate); err != nil {
		return errors.ApiParamValidError(err)
	}

	if err := clickCreate.Validate(ctx); err != nil {
		log.Errorf("Create Error:  %s", err.Error())
		return err
	}

	err := service.ClickService().Create(ctx.Request().Context(), clickCreate)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

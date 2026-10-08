package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (a *API) GetHolidays(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	if year == 0 {
		year = time.Now().Year()
	}
	if a.holidays == nil {
		BadRequest(c, "节假日服务未启用")
		return
	}
	y, err := a.holidays.Year(c.Request.Context(), year)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, y)
}

func (a *API) RefreshHolidays(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	if year == 0 {
		year = time.Now().Year()
	}
	if a.holidays == nil {
		BadRequest(c, "节假日服务未启用")
		return
	}
	y, err := a.holidays.RefreshYear(c.Request.Context(), year, "manual")
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, y)
}

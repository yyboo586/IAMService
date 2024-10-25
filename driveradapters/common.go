package driveradapters

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

func validate(c *gin.Context) (i interface{}, err error) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	return i, json.Unmarshal(data, &i)
}

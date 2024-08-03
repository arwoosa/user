package helpers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Validate(c *gin.Context, arr interface{}) error {
	errorList := []string{}
	if err := c.BindJSON(&arr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -2, "message": "Validation error! JSON does not match", "data": errorList, "validation": "oosa_api"})
		return err
	}

	validate := validator.New()

	err := validate.Struct(arr)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, e := range validationErrors {
			//translatedErr := fmt.Errorf(e.Translate(trans))
			//errs = append(errs, translatedErr)
			errorList = append(errorList, e.Error())
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": -2, "message": "Validation error! Please check your inputs!", "data": errorList, "validation": "oosa_api"})
		return err
	}

	return nil
}

func ValidateWithShouldBind(c *gin.Context, obj interface{}) error {
	errorList := []string{}
	if err := c.ShouldBind(obj); err != nil {
		fmt.Printf("error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": -2, "message": "Validation error! Request data not complete", "data": errorList, "validation": "oosa_api"})
		return err
	}

	validate := validator.New()

	err := validate.Struct(obj)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, e := range validationErrors {
			errorList = append(errorList, e.Error())
		}
		c.JSON(http.StatusOK, gin.H{"code": -2, "message": "Validation error! Please check your inputs!", "data": errorList, "validation": "oosa_api"})
		return err
	}

	return nil
}

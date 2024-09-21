package auth

import (
	"ManyACG/common"
	. "ManyACG/logger"
	"ManyACG/model"
	"ManyACG/service"
	"ManyACG/types"
	"errors"
	"net/http"
	"regexp"

	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type SendCodeRequest struct {
	Username   string `json:"username" form:"username" binding:"required,min=4,max=20"`
	AuthMethod string `json:"auth_method" form:"auth_method" binding:"required,oneof=telegram"`
}

func handleSendCode(c *gin.Context) {
	var request SendCodeRequest
	if err := c.ShouldBind(&request); err != nil {
		common.GinBindError(c, err)
		return
	}

	if !regexp.MustCompile("^[a-zA-Z0-9_]+$").MatchString(request.Username) {
		c.JSON(http.StatusBadRequest, common.RestfulCommonResponse[any]{Status: http.StatusBadRequest, Message: "username must be alphanumeric"})
		return
	}

	user, err := service.GetUserByUsername(c, request.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		Logger.Errorf("Failed to get user: %v", err)
		common.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to get user")
		return
	}
	if user != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  http.StatusConflict,
			"message": "User already exists",
		})
		return
	}
	unauthUserInDB, err := service.GetUnauthUserByUsername(c, request.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		Logger.Errorf("Failed to get user: %v", err)
		common.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to get user")
		return
	}
	if unauthUserInDB != nil {
		common.GinErrorResponse(c, err, http.StatusConflict, "User already exists")
		return
	}
	authMethod := types.AuthMethod(request.AuthMethod)
	code := random.RandNumeral(6)
	unauthUser, err := service.CreateUnauthUser(c, &model.UnauthUserModel{
		Username:   request.Username,
		AuthMethod: authMethod,
		Code:       code,
	})
	if err != nil {
		Logger.Errorf("Failed to create unauth user: %v", err)
		common.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed send code")
		return
	}

	// Telegram auth method needs user to /start the bot to get the code
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Code sent",
		"data": gin.H{
			"id": unauthUser.ID.Hex(), // ID of the unauth user, used to generate the deep link
		},
	})
}

type RegisterRequest struct {
	Username   string `json:"username" form:"username" binding:"required,min=4,max=20" msg:"Username must be between 4 and 20 characters"`
	Password   string `json:"password" form:"password" binding:"required,min=6,max=32" msg:"Password must be between 6 and 32 characters"`
	AuthMethod string `json:"auth_method" form:"auth_method" binding:"required,oneof=telegram" msg:"Auth method now only supports telegram"`
	Code       string `json:"code" form:"code" binding:"required,min=6,max=6" msg:"Code must be 6 characters"`
	TelegramID int64  `json:"telegram_id" form:"telegram_id"`
	// Email      string `json:"email" form:"email" binding:"omitempty,email" msg:"Invalid email"`
}

func handleRegister(c *gin.Context) {
	var register RegisterRequest
	if err := c.ShouldBind(&register); err != nil {
		common.GinBindError(c, err)
		return
	}

	if !regexp.MustCompile("^[a-zA-Z0-9_]+$").MatchString(register.Username) {
		c.JSON(http.StatusBadRequest, common.RestfulCommonResponse[any]{Status: http.StatusBadRequest, Message: "username must be alphanumeric"})
		return
	}

	user, err := service.GetUserByUsername(c, register.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		Logger.Errorf("Failed to get user: %v", err)
		common.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to get user")
		return
	}
	if user != nil {
		common.GinErrorResponse(c, err, http.StatusConflict, "User already exists")
		return
	}

	unauthUser, err := service.GetUnauthUserByUsername(c, register.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		Logger.Errorf("Failed to get unauth user: %v", err)
		common.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to verify code")
		return
	}
	if unauthUser == nil {
		c.JSON(http.StatusNotFound, common.RestfulCommonResponse[any]{Status: http.StatusNotFound, Message: "User not found"})
		return
	}
	if unauthUser.Code != register.Code {
		c.JSON(http.StatusUnauthorized, common.RestfulCommonResponse[any]{Status: http.StatusUnauthorized, Message: "Invalid code"})
		return
	}
	if unauthUser.AuthMethod == types.AuthMethodTelegram && unauthUser.TelegramID != register.TelegramID {
		c.JSON(http.StatusUnauthorized, common.RestfulCommonResponse[any]{Status: http.StatusUnauthorized, Message: "Telegram ID does not match"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)
	if err != nil {
		Logger.Errorf("Failed to hash password: %v", err)
		common.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	_, err = service.CreateUser(c, &model.UserModel{
		Username:   register.Username,
		Password:   string(hashedPassword),
		TelegramID: register.TelegramID,
	})
	if err != nil {
		Logger.Errorf("Failed to create user: %v", err)
		common.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to create user")
		return
	}
	c.JSON(http.StatusOK, common.RestfulCommonResponse[any]{Status: http.StatusOK, Message: "User created"})
	if err := service.DeleteUnauthUser(c, unauthUser.ID); err != nil {
		Logger.Warnf("Failed to delete unauth user: %v", err)
	}
}

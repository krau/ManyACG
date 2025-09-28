package auth

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/internal/infra/config"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/types"

	"github.com/duke-git/lancet/v2/random"
	lancetValidator "github.com/duke-git/lancet/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type SendCodeRequest struct {
	Username   string `json:"username" form:"username" binding:"required,min=4,max=20"`
	AuthMethod string `json:"auth_method" form:"auth_method" binding:"required,oneof=telegram email"`
	Email      string `json:"email" form:"email" binding:"omitempty,email"`
}

func handleSendCode(c *gin.Context) {
	var request SendCodeRequest
	if err := c.ShouldBind(&request); err != nil {
		utils.GinBindError(c, err)
		return
	}

	if !regexp.MustCompile("^[a-zA-Z0-9_]+$").MatchString(request.Username) {
		c.JSON(http.StatusBadRequest, utils.RestfulCommonResponse[any]{Status: http.StatusBadRequest, Message: "username must be alphanumeric"})
		return
	}

	user, err := service.GetUserByUsername(c, request.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		common.Logger.Errorf("Failed to get user: %v", err)
		utils.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to get user")
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
		common.Logger.Errorf("Failed to get user: %v", err)
		utils.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to get user")
		return
	}
	if unauthUserInDB != nil {
		utils.GinErrorResponse(c, errors.New("user already exists"), http.StatusConflict, "User already exists")
		return
	}
	authMethod := types.AuthMethod(request.AuthMethod)

	// TODO: refactor this
	if authMethod == types.AuthMethodEmail {
		if request.Email == "" {
			c.JSON(http.StatusBadRequest, utils.RestfulCommonResponse[any]{Status: http.StatusBadRequest, Message: "Email is required"})
			return
		}
		if !lancetValidator.IsEmail(request.Email) {
			c.JSON(http.StatusBadRequest, utils.RestfulCommonResponse[any]{Status: http.StatusBadRequest, Message: "Invalid email"})
			return
		}
		if _, err := service.GetUserByEmail(c, request.Email); err == nil {
			c.JSON(http.StatusConflict, utils.RestfulCommonResponse[any]{Status: http.StatusConflict, Message: "Email already exists"})
			return
		}
		if common.ResendClient == nil {
			c.JSON(http.StatusInternalServerError, utils.RestfulCommonResponse[any]{Status: http.StatusInternalServerError, Message: "Sorry, not supported yet"})
			return
		}
	}

	code := random.RandNumeral(6)
	unauthUser, err := service.CreateUnauthUser(c, &types.UnauthUserModel{
		Username:   request.Username,
		AuthMethod: authMethod,
		Email:      request.Email,
		Code:       code,
	})
	if err != nil {
		common.Logger.Errorf("Failed to create unauth user: %v", err)
		utils.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed send code")
		return
	}

	switch authMethod {
	case types.AuthMethodTelegram:
		// Telegram auth method needs user to /start the bot to get the code
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Code generated",
			"data": gin.H{
				"id": unauthUser.ID.Hex(), // ID of the unauth user, used to generate the deep link
			},
		})
	case types.AuthMethodEmail:
		_, err = common.ResendClient.Emails.Send(&resend.SendEmailRequest{
			From:    config.Get().Auth.Resend.From,
			To:      []string{request.Email},
			Subject: config.Get().Auth.Resend.Subject,
			Text:    "你的验证码是: " + code + ".\n\n请在 10 分钟内使用, 请勿泄露给他人",
		})
		if err != nil {
			common.Logger.Errorf("Failed to send email: %v", err)
			utils.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to send email")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Code sent to <" + request.Email + ">",
		})
	}
}

type RegisterRequest struct {
	Username   string `json:"username" form:"username" binding:"required,min=4,max=20" msg:"Username must be between 4 and 20 characters"`
	Password   string `json:"password" form:"password" binding:"required,min=6,max=32" msg:"Password must be between 6 and 32 characters"`
	AuthMethod string `json:"auth_method" form:"auth_method" binding:"required,oneof=telegram email" msg:"Auth method now only supports telegram"`
	Code       string `json:"code" form:"code" binding:"required,min=6,max=6" msg:"Code must be 6 characters"`
	TelegramID int64  `json:"telegram_id" form:"telegram_id" binding:"omitempty" msg:"Invalid telegram ID"`
	Email      string `json:"email" form:"email" binding:"omitempty,email" msg:"Invalid email"`
}

func handleRegister(c *gin.Context) {
	var register RegisterRequest
	if err := c.ShouldBind(&register); err != nil {
		utils.GinBindError(c, err)
		return
	}

	if !regexp.MustCompile("^[a-zA-Z0-9_]+$").MatchString(register.Username) {
		c.JSON(http.StatusBadRequest, utils.RestfulCommonResponse[any]{Status: http.StatusBadRequest, Message: "username must be alphanumeric"})
		return
	}

	user, err := service.GetUserByUsername(c, register.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		common.Logger.Errorf("Failed to get user: %v", err)
		utils.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to get user")
		return
	}
	if user != nil {
		utils.GinErrorResponse(c, errors.New("user already exists"), http.StatusConflict, "User already exists")
		return
	}

	unauthUser, err := service.GetUnauthUserByUsername(c, register.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		common.Logger.Errorf("Failed to get unauth user: %v", err)
		utils.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to verify code")
		return
	}
	if unauthUser == nil {
		c.JSON(http.StatusNotFound, utils.RestfulCommonResponse[any]{Status: http.StatusNotFound, Message: "User not found"})
		return
	}
	if unauthUser.Code != register.Code {
		c.JSON(http.StatusUnauthorized, utils.RestfulCommonResponse[any]{Status: http.StatusUnauthorized, Message: "Invalid code"})
		return
	}
	if unauthUser.AuthMethod == types.AuthMethodTelegram && unauthUser.TelegramID != register.TelegramID {
		c.JSON(http.StatusUnauthorized, utils.RestfulCommonResponse[any]{Status: http.StatusUnauthorized, Message: "Telegram ID does not match"})
		return
	}
	if unauthUser.AuthMethod == types.AuthMethodEmail {
		if unauthUser.Email != register.Email {
			c.JSON(http.StatusUnauthorized, utils.RestfulCommonResponse[any]{Status: http.StatusUnauthorized, Message: "Email does not match"})
			return
		}
		if _, err := service.GetUserByEmail(c, register.Email); err == nil {
			c.JSON(http.StatusConflict, utils.RestfulCommonResponse[any]{Status: http.StatusConflict, Message: "Email already exists"})
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)
	if err != nil {
		common.Logger.Errorf("Failed to hash password: %v", err)
		utils.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	_, err = service.CreateUser(c, &types.UserModel{
		Username:   register.Username,
		Password:   string(hashedPassword),
		TelegramID: register.TelegramID,
		Email:      register.Email,
	})
	if err != nil {
		common.Logger.Errorf("Failed to create user: %v", err)
		utils.GinErrorResponse(c, err, http.StatusInternalServerError, "Failed to create user")
		return
	}
	c.JSON(http.StatusOK, utils.RestfulCommonResponse[any]{Status: http.StatusOK, Message: "User created"})
	if err := service.DeleteUnauthUser(c, unauthUser.ID); err != nil {
		common.Logger.Warnf("Failed to delete unauth user: %v", err)
	}
}

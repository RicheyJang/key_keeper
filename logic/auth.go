package logic

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/RicheyJang/key_keeper/model"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	orgjwt "github.com/kataras/jwt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// UserLoginRequest is the request struct for user login
type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserClaims struct {
	ID    uint
	Name  string
	Level int
}

// Validate UserClaims验证函数，将在JWT Verify时调用
func (uc UserClaims) Validate() error {
	m := GetManager()
	if _, ok := m.frozenUsers.Load(uc.ID); ok { // 检查该用户是否被冻结
		return errors.UserFrozen
	}
	return nil
}

// GetLoginHandler 获取登录处理函数
func (manager *Manager) GetLoginHandler() iris.Handler {
	secret := get256SecretKey()
	maxAge := viper.GetDuration("user.maxAge")
	if maxAge < time.Minute {
		maxAge = time.Minute
	}

	signer := jwt.NewSigner(jwt.HS256, secret, maxAge)
	return func(ctx iris.Context) {
		// 绑定请求
		var req UserLoginRequest
		if err := ctx.ReadJSON(&req); err != nil {
			responseError(ctx, errors.InvalidRequest)
			return
		}
		// 校验
		user, err := manager.userManager.CheckUser(req.Username, req.Password)
		if err != nil {
			responseError(ctx, errors.WrongPasswd)
			return
		}
		if user.IsFrozen {
			responseError(ctx, errors.UserFrozen)
			return
		}
		// 生成token
		claims := UserClaims{
			ID:    user.ID,
			Name:  user.Name,
			Level: user.Level,
		}
		token, err := signer.Sign(claims)
		if err != nil {
			log.Errorf("sign token error: %v", err)
			responseError(ctx, errors.Unknown)
			return
		}
		// 记录用户登录时间、登录IP
		_ = manager.userManager.SaveUserLoginInfo(model.User{
			ID:        user.ID,
			LastIP:    ctx.RemoteAddr(),
			LastLogin: time.Now(),
		})
		responseSuccess(ctx, "data", iris.Map{
			"token":    string(token),
			"username": user.Name,
		})
	}
}

// GetVerifyHandler 获取验证处理函数
func (manager *Manager) GetVerifyHandler() iris.Handler {
	secret := get256SecretKey()
	maxAge := viper.GetDuration("user.maxAge")
	if maxAge < time.Minute {
		maxAge = time.Minute
	}

	// 创建验证器
	verifier := jwt.NewVerifier(jwt.HS256, secret)
	verifier.Blocklist = orgjwt.NewBlocklist(maxAge)
	verifier.Extractors = []jwt.TokenExtractor{jwt.FromHeader} // extract token only from Authorization: Bearer $token
	verifier.ErrorHandler = func(ctx iris.Context, err error) {
		responseError(ctx, errors.InvalidToken)
	}
	return verifier.Verify(func() interface{} {
		return new(UserClaims)
	})
}

// HandlerOfLogout 注销处理函数
func (manager *Manager) HandlerOfLogout(ctx iris.Context) {
	_ = ctx.Logout()
	responseSuccess(ctx, "", nil)
}

// 获取用户token内信息
func (manager *Manager) getUserClaims(ctx iris.Context) *UserClaims {
	claims, ok := jwt.Get(ctx).(*UserClaims)
	if !ok || claims == nil {
		return new(UserClaims)
	}
	return claims
}

// 获取用户token内信息
func (manager *Manager) getUserID(ctx iris.Context) uint {
	claims, ok := jwt.Get(ctx).(*UserClaims)
	if !ok || claims == nil {
		return 0
	}
	return claims.ID
}

var jwtSecretKey []byte
var jwtSecretKeyOnce sync.Once

func get256SecretKey() []byte {
	jwtSecretKeyOnce.Do(func() {
		b := make([]byte, 32)
		l, err := rand.Read(b)
		if l != 32 || err != nil {
			b = []byte("sercreThatmaycontainch@r$32charS")
		}
		jwtSecretKey = b
	})
	return jwtSecretKey
}

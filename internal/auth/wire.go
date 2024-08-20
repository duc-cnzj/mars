package auth

//go:generate mockgen -destination ./mock_auth.go -package auth github.com/duc-cnzj/mars/v4/internal/auth Auth
import "github.com/google/wire"

var WireAuth = wire.NewSet(NewAuthn)

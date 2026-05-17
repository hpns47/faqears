package handler

import (
	"context"

	"github.com/faqears/faqears/gateway/internal/middleware"
	"github.com/faqears/faqears/pkg/grpcx"
)

func callerIdentity(ctx context.Context) grpcx.Identity {
	roles, _ := ctx.Value(middleware.UserRolesKey).([]string)
	return grpcx.Identity{
		UserID: getString(ctx, middleware.UserIDKey),
		Email:  getString(ctx, middleware.UserEmailKey),
		Roles:  roles,
	}
}

func getString(ctx context.Context, key interface{}) string {
	v, _ := ctx.Value(key).(string)
	return v
}

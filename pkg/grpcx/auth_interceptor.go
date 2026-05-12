package grpcx

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	MDUserID = "x-faqears-user-id"
	MDRoles  = "x-faqears-user-roles"
	MDEmail  = "x-faqears-user-email"
)

type identityCtxKey struct{}

type Identity struct {
	UserID string
	Email  string
	Roles  []string
}

func IdentityFromContext(ctx context.Context) Identity {
	if v, ok := ctx.Value(identityCtxKey{}).(Identity); ok {
		return v
	}
	return Identity{}
}

func UserIDFromCtx(ctx context.Context) string {
	return IdentityFromContext(ctx).UserID
}

func RolesFromCtx(ctx context.Context) []string {
	return IdentityFromContext(ctx).Roles
}

func HasRole(ctx context.Context, role string) bool {
	for _, r := range RolesFromCtx(ctx) {
		if r == role {
			return true
		}
	}
	return false
}

func IdentityInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		id := Identity{
			UserID: first(md, MDUserID),
			Email:  first(md, MDEmail),
			Roles:  parseRoles(first(md, MDRoles)),
		}
		ctx = context.WithValue(ctx, identityCtxKey{}, id)
		return handler(ctx, req)
	}
}

func RequireRole(role string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !HasRole(ctx, role) {
			return nil, status.Error(codes.PermissionDenied, "missing required role: "+role)
		}
		return handler(ctx, req)
	}
}

func RequireAnyRole(roles ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		for _, r := range roles {
			if HasRole(ctx, r) {
				return handler(ctx, req)
			}
		}
		return nil, status.Error(codes.PermissionDenied, "missing required role")
	}
}

func RequireMethodRole(roleByMethod map[string]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		role, ok := roleByMethod[info.FullMethod]
		if !ok {
			return handler(ctx, req)
		}
		if !HasRole(ctx, role) {
			return nil, status.Error(codes.PermissionDenied, "missing required role for "+info.FullMethod+": "+role)
		}
		return handler(ctx, req)
	}
}

func RequireMethodAnyRole(rolesByMethod map[string][]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		roles, ok := rolesByMethod[info.FullMethod]
		if !ok {
			return handler(ctx, req)
		}
		for _, r := range roles {
			if HasRole(ctx, r) {
				return handler(ctx, req)
			}
		}
		return nil, status.Error(codes.PermissionDenied, "missing required role for "+info.FullMethod)
	}
}

func first(md metadata.MD, key string) string {
	v := md.Get(key)
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func parseRoles(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func OutgoingIdentity(ctx context.Context, id Identity) context.Context {
	md := metadata.New(map[string]string{
		MDUserID: id.UserID,
		MDEmail:  id.Email,
		MDRoles:  strings.Join(id.Roles, ","),
	})
	return metadata.NewOutgoingContext(ctx, md)
}

package ctxutil

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func GetLocale(ctx context.Context) string {
	locale := GetValueFromContext(ctx, localeKey)
	if len(locale) == 0 {
		return ""
	}
	return locale[0]
}

func SetLocale(ctx context.Context, locale string) context.Context {
	ctx = context.WithValue(ctx, localeKey, []string{locale})
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md[localeKey] = []string{locale}
	return metadata.NewOutgoingContext(ctx, md)
}

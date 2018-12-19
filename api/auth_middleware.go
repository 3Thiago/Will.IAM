package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/ghostec/Will.IAM/usecases"
)

type serviceAccountIDCtxKeyType string

const serviceAccountIDCtxKey = serviceAccountIDCtxKeyType("serviceAccountID")

func getServiceAccountID(ctx context.Context) (string, bool) {
	v := ctx.Value(serviceAccountIDCtxKey)
	vv, ok := v.(string)
	if !ok {
		return "", false
	}
	return vv, true
}

// authMiddleware authenticates either access_token or key pair
func authMiddleware(
	saUseCase usecases.ServiceAccounts,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("authorization")
			parts := strings.Split(authorization, " ")
			if authorization == "" || len(parts) != 2 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			var ctx context.Context
			if parts[0] == "KeyPair" {
				keyPair := strings.Split(parts[1], ":")
				saID, err := saUseCase.AuthenticateKeyPair(keyPair[0], keyPair[1])
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				ctx = context.WithValue(r.Context(), serviceAccountIDCtxKey, saID)
			} else if parts[0] == "Bearer" {
				accessToken := parts[1]
				accessTokenAuth, err := saUseCase.AuthenticateAccessToken(accessToken)
				if err != nil {
					println(err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("x-access-token", accessTokenAuth.AccessToken)
				ctx = context.WithValue(
					r.Context(), serviceAccountIDCtxKey, accessTokenAuth.ServiceAccountID,
				)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

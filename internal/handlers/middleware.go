package handlers

import (
	"context"
	"net/http"

	"poker-planning/internal/models"
	"poker-planning/internal/services"
)

const (
	UserContextKey = "user"
	SessionCookieName = "poker_session"
)

func SessionMiddleware(userService *services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := userService.GetUserByID(cookie.Value)
			if err != nil {
				http.SetCookie(w, &http.Cookie{
					Name:     SessionCookieName,
					Value:    "",
					MaxAge:   -1,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				})
				next.ServeHTTP(w, r)
				return
			}

			if user == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     SessionCookieName,
					Value:    "",
					MaxAge:   -1,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				})
				next.ServeHTTP(w, r)
				return
			}

			userService.UpdateLastSeen(user.ID)

			http.SetCookie(w, &http.Cookie{
				Name:     SessionCookieName,
				Value:    user.ID,
				MaxAge:   6 * 3600, // 6 hours
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// JWTMiddleware проверяет наличие и валидность JWT-токена в заголовке Authorization.
// Если токен валиден, он извлекает ID пользователя и прокидывает его дальше в контексте запроса.
func JWTMiddleware(jwtSecret string, logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Получаем заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			// 2. Проверяем формат заголовка (должен быть "Bearer <токен>")
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format. Use 'Bearer <token>'")
				return
			}

			tokenStr := parts[1]
			claims := &jwt.RegisteredClaims{}

			// 3. Парсим и проверяем подпись токена с помощью нашего секретного ключа
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				// Проверяем метод подписи (должен быть HMAC HS256)
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			// Если токен просрочен, подделан или невалиден
			if err != nil || !token.Valid {
				logger.WithField("token_err", err).Warn("Unauthorized access attempt: invalid token")
				respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// 4. Извлекаем ID пользователя (он хранится в поле Subject (sub) внутри claims)
			userID := claims.Subject
			if userID == "" {
				respondWithError(w, http.StatusUnauthorized, "Invalid token claims: user ID missing")
				return
			}

			// 5. Ключевой момент: кладем userID в контекст текущего HTTP-запроса.
			// Благодаря этому любой хэндлер (например, перевод денег) сможет узнать, какой именно пользователь делает запрос.
			ctx := context.WithValue(r.Context(), "userID", userID)

			// Передаем управление следующему middleware или хэндлеру, используя обновленный контекст
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Вспомогательная функция для отправки JSON-ошибок (чтобы не дублировать код)
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

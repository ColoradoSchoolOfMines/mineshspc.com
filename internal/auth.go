package internal

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

func (a *Application) GetLoggedInTeacher(r *http.Request) (*database.Teacher, error) {
	jwtStr, err := r.Cookie("tok")
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(jwtStr.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return a.Config.ReadSecretKey(), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !token.Valid || !ok {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.Issuer != string(IssuerSessionToken) {
		return nil, fmt.Errorf("token is not a session token")
	}

	user, err := a.DB.GetTeacherByEmail(r.Context(), claims.Subject)
	if err != nil {
		a.Log.Warn().Err(err).
			Interface("claims", claims).
			Msg("couldn't find teacher with that session token")
		return nil, err
	}

	return user, nil
}

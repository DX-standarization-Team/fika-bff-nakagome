package middleware

import (
	"log"
	"net/http"
	"net/url"

	"strings"
	"time"

	"github.com/pkg/errors"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

const AUTH0_AUDIENCE = "https://fs-apigw-bff-nakagome-bi5axj14.uc.gateway.dev/"
const AUTH0_DOMAIN = "dev-kjqwuq76z8suldgw.us.auth0.com"

// // CustomClaims contains custom data we want from the token.
// type CustomClaims struct {
// 	Scope string `json:"scope"`
// }

// // Validate does nothing for this example, but we need
// // it to satisfy validator.CustomClaims interface.
// func (c CustomClaims) Validate(ctx context.Context) error {
// 	return nil
// }

// scopeは今回使用しない
// // HasScope checks whether our claims have a specific scope.
// func (c CustomClaims) HasScope(expectedScope string) bool {
// 	result := strings.Split(c.Scope, " ")
// 	for i := range result {
// 		if result[i] == expectedScope {
// 			return true
// 		}
// 	}

// 	return false
// }

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() func(next http.Handler) http.Handler {
	// issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	issuerURL, err := url.Parse("https://" + AUTH0_DOMAIN + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}
	log.Printf("Parsed issuerURL: %s", issuerURL)

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	log.Printf("provider: %v", provider)
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{AUTH0_AUDIENCE},
		// []string{os.Getenv("AUTH0_AUDIENCE")},
		// 今回はカスタムクレーム特にみない
		// validator.WithCustomClaims(
		// 	func() validator.CustomClaims {
		// 		return &CustomClaims{}
		// 	},
		// ),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}
	log.Printf("jwtValidator: %v", jwtValidator)

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	log.Printf("create jwtMiddleware instance")
	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
		jwtmiddleware.WithTokenExtractor(AuthHeaderTokenExtractor),
	)

	return func(next http.Handler) http.Handler {
		log.Printf("middleware.CheckJWT starts")
		return middleware.CheckJWT(next)
	}
}

func AuthHeaderTokenExtractor(r *http.Request) (string, error) {
	log.Printf("自作AuthHeaderTokenExtractor")
	// authHeader := r.Header.Get("Authorization")
	authHeader := r.Header.Get("X-Forwarded-Authorization")
	if authHeader == "" {
		return "", nil // No error, just no JWT.
	}

	authHeaderParts := strings.Fields(authHeader)
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", errors.New("Authorization header format must be Bearer {token}")
	}
	log.Printf("自作AuthHeaderTokenExtractor exiting : %v", authHeaderParts[1])
	return authHeaderParts[1], nil
}

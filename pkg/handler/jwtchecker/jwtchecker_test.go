// Package jwtchecker is the package that contains the check functions for JWT.
package jwtchecker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"github.com/MicahParks/jwkset"
	"github.com/stretchr/testify/assert"
)

// errTokenUnverifiable is an error that occurs when the JWT is unverifiable.
var errTokenUnverifiable = errors.New("token is unverifiable: error while executing keyfunc: key not found: " +
	"kid \"9LbeNC5ZmVA9RXwHSeBqCCugdtgbfM2pa5VHz-PGPV0\"\n" +
	"failed keyfunc: could not read JWK from storage")

const (
	// validJWT1 is the first valid JWT used for testing.
	//
	// The JWT is obtained from the public sources and should not be trusted, nor is it a valid JWT for any cloud.
	validJWT1 = "eyJhbGciOiJSUzI1NiIsImtpZCI6IkNuUnpSVmJHdEtqTWFzenJrUzNOeExUVmY0SVZSZE5vcDZ2bEdMMUpNMFEiLCJ0eXAiOiJKV1QifQ." +
		"eyJhdWQiOlsidmFsaWQtMSJdLCJzdWIiOiJhbHBoYS1zZW5zZS5jb20ifQ.Op7O9gpRw9dQxStghGoh8gZ3A7DWAwW8n5cGOI57oCzhFgro" +
		"5KjcFKaMsMnUSDi5zU_ngU1DI9HIsup5MSyoJBzXK1THB0ziu-3F5xSnwU1iKQNnb898wEd-7srboY_G86-G0RAFFM38WVHO-iifEXGGBgH" +
		"1D6inN9IO_jeM7H6MvkgtTKwvyJ-Q4c1CTiWRacY4MhrIcFVCcKJ0u1csRRklDZanz9YEfIddRayIYV3dL9Agwfz1Sa-5MOQJQdgiYQIPyQ" +
		"mTcap2W6OSnAeY92YX09Uemi3TWHNWEi1pitNcQkGAwiMqKEX2SMtrl6aP1WhPYv--OgY0YiR5zot3wQ"

	// validJWT2 is the second valid JWT used for testing.
	//
	// The JWT is obtained from the public sources and should not be trusted, nor is it a valid JWT for any cloud.
	validJWT2 = "eyJhbGciOiJSUzI1NiIsImtpZCI6IkNuUnpSVmJHdEtqTWFzenJrUzNOeExUVmY0SVZSZE5vcDZ2bEdMMUpNMFEiLCJ0eXAiOiJKV1QifQ." +
		"eyJhdWQiOlsidmFsaWQtMiJdLCJzdWIiOiJhbHBoYS1zZW5zZS5jb20ifQ.lDZAn9vsDZuPEu07Rhahw7lon9pAyBjXL0QJjsqdK7A5R7kQ" +
		"H46D5ojejFk-GCtHt9xrSZpH6NC8UcRpm0FHdTdeFA7q8Us0iiZTp13EHA3UhBOsMQi3udf5ZIKTVk4pBU0ApTAmtu64WH5K5hZ62PhLc16" +
		"i3QBGzJPFbEZMOycONG8CaFkxuCCABnaK4dEFqrOBxwJX7CR05P8IVCXNS3P2sQZ9bkwpfIvdzMKPSqg-UpPgU_Ef0SbSQ2iYgTrFzwHxOq" +
		"_xVlhMP5p2cvI_kSzSWZwqIZFXgclkwtJhsHxpQUF_x1fpefkrDXut4o2c5bLiMYGRH0nx81XsBSrJOQ"

	// invalidJWT is an invalid JWT used for testing.
	//
	// The JWT is obtained from the public sources and should not be trusted, nor is it a valid JWT for any cloud.
	invalidJWT = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjlMYmVOQzVabVZBOVJYd0hTZUJxQ0N1Z2R0Z2JmTTJwYTVWSHotUEdQVjAiLCJ0eXAiOiJKV1QifQ." +
		"eyJhdWQiOlsiaW52YWxpZCJdLCJzdWIiOiJhbHBoYS1zZW5zZS5jb20ifQ.YgJTXY5eBdEJ10xyyzQQlGp-sNxalbnZntGqX-eCsCXBeOAU" +
		"csVODXVzLFnRDDSzV-Ug3ToEuFWgOjbqdksqAOZQ_X3169FB67r6kRPDbghG5dL9znTDBl3RoWMyrehImfoIoGmxQeV3VluvoXKYYPLdKaB" +
		"OCcnRvob2h4lwxr6s4IfnZVK6iJxJ_Dnx7kVIn0r4RIo6tuC74q7yuICwmRc8LIjfoYqS9TdR0XuiUexgvxhDEMpR8ZN_OZrHY1UqSdIOK9" +
		"ATtix1F8aW-fp5A7jP-MRoYGWdVKBsy7V7wOfDfi3_myFc-_ylmENDs4qLkNXm_OXta5lCSLpz0AuhyA"
)

// constJWKsJSON is a JWKS JSON used for testing.
//
// The JWKS JSON is obtained from the public sources and should not be trusted, nor is it a valid JWKS JSON of any cloud.
// Do not modify this variable, it is supposed to be constant.
var constJWKsJSON = jwkset.JWKSMarshal{
	Keys: []jwkset.JWKMarshal{{
		ALG: jwkset.AlgRS256,
		E:   "AQAB",
		KID: "CnRzRVbGtKjMaszrkS3NxLTVf4IVRdNop6vlGL1JM0Q",
		KTY: jwkset.KtyRSA,
		N: "zppahJ5jtkM3UFr0--m-auZ6D8fHHCPdHYJW72jR0wPg58g3pEBDTfKX-b9s354xhV1_wEzaA4meXPXiFCrJNJ6ySxRwo4UtVhfH8h5DrCn" +
			"_U9Zf4LhXfQ6ulDaEung8o9zlVT1LohhjAAwtJBMVvezq8m8ARJlG2z4jY3Sz7FG1vWOkO-OkjPOLgZ_4Im9SBkc5oMyIVRyksvSSfqKPbX" +
			"1U3ReZ0imkv-16Mefulqyt_HWYzRibyPR7Woq7-z4TAGU7tjuKf-YOqDvVTMH_1luaBJW7moOKqJlFtMMoxOtyZ5v0k-lQOXdLelW7kgSma" +
			"_ppbkxpdJB_wq3OvI1RvQ",
		USE: jwkset.UseSig,
	}},
}

// setupJWTCheckerTest creates a new JWTChecker instance for testing.
func setupJWTCheckerTest() (*JWTChecker, *string) {
	mockHTTPServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(constJWKsJSON)
	}))

	return New(mockHTTPServer.Client()), &mockHTTPServer.URL
}

// TestJWTChecker_Check tests the Check method of the JWTChecker.
func TestJWTChecker_Check(t *testing.T) {
	testCases := []struct {
		name    string
		jwts    []*string
		wantErr error
	}{
		{
			name:    "One valid JWT",
			jwts:    []*string{util.Ref(validJWT1)},
			wantErr: nil,
		},
		{
			name:    "Two valid JWTs",
			jwts:    []*string{util.Ref(validJWT1), util.Ref(validJWT2)},
			wantErr: nil,
		},
		{
			name:    "One valid JWT, one invalid JWT",
			jwts:    []*string{util.Ref(validJWT1), util.Ref(invalidJWT)},
			wantErr: errTokenUnverifiable,
		},
		{
			name:    "One invalid JWT",
			jwts:    []*string{util.Ref(invalidJWT)},
			wantErr: errTokenUnverifiable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jwtChecker, jwksURI := setupJWTCheckerTest()

			_, gotErr := jwtChecker.Handle(context.TODO(), jwksURI, tc.jwts)

			if tc.wantErr != nil {
				assert.Error(t, gotErr, "expected error %v, got %v", tc.wantErr, gotErr)
			}
		})
	}
}

package authenticator

import (
	"log"
	"net/http"

	"github.com/abbot/go-http-auth"
	"github.com/fxnn/gone/context"
	"github.com/fxnn/gone/router"
)

const (
	authenticationRealmName = "gone wiki"
)

type HttpBasicAuthenticator struct {
	authenticationHandler *auth.BasicAuth
	authenticationStore   cookieAuthenticationStore
}

func NewHttpBasicAuthenticator() *HttpBasicAuthenticator {
	var authenticationHandler = auth.NewBasicAuthenticator(authenticationRealmName, provideSampleSecret)
	var authenticationStore = newCookieAuthenticationStore()
	return &HttpBasicAuthenticator{authenticationHandler, authenticationStore}
}

// AuthHandler wraps an http.Handler and reads authentication information from
// the session associated with the request prior to calling the delegate handler.
func (a *HttpBasicAuthenticator) AuthHandler(delegate http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if userId, ok := a.authenticationStore.userId(request); ok {
			a.setUserId(request, userId)
		}

		delegate.ServeHTTP(writer, request)
	})
}

// ServeHTTP serves an authentication UI.
func (a *HttpBasicAuthenticator) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	a.checkAuth(request)

	if a.IsAuthenticated(request) {
		log.Printf("%s %s: authenticated as %s", request.Method, request.URL, a.UserId(request))
		a.authenticationStore.setUserId(writer, request, a.UserId(request))
		router.RedirectToViewMode(writer, request)
		return
	}

	a.authenticationHandler.RequireAuth(writer, request)
}

func (a *HttpBasicAuthenticator) checkAuth(request *http.Request) {
	a.setUserId(request, a.authenticationHandler.CheckAuth(request))
}

func (a *HttpBasicAuthenticator) IsAuthenticated(request *http.Request) bool {
	return context.Load(request).IsAuthenticated()
}

func (a *HttpBasicAuthenticator) UserId(request *http.Request) string {
	return context.Load(request).UserId
}

func (a *HttpBasicAuthenticator) setUserId(request *http.Request, userId string) {
	var ctx = context.Load(request)
	ctx.UserId = userId
	ctx.Save(request)
}

func provideSampleSecret(user, realm string) string {
	if user == "test" {
		// password is "hello"
		return "$1$dlPL2MqE$oQmn16q49SqdmhenQuNgs1"
	}
	return ""
}
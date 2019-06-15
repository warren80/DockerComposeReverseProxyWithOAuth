package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/securecookie"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/unrolled/render.v1"
	"log"
	"net/http"
	"time"
)

var (
	nsCookieName         = "NSLOGIN"
	nsCookieHashKey      = []byte("SECURE_COOKIE_HASH_KEY")
	nsRedirectCookieName = "NSREDIRECT"
	cfg                  config
)

type config struct {
	// e.g. https://protected-resource.example.redbyte.eu
	DefaultRedirectUrl string
	// shared password
	Password string
	// shared domain prefix between protected resource and auth server
	// e.g. .example.redbyte.eu (note the leading dot)
	Domain string
}

func main() {

	// configuration
	port := flag.Int("port", 8888, "listen port")
	flag.Parse()
	var err error
	if cfg, err = loadConfig("config.toml"); err != nil {
		log.Fatal(err)
	}

	// template renderer
	rndr := render.New(render.Options{
		Directory:     "templates",
		IsDevelopment: false,
	})

	// router
	router := httprouter.New()
	router.GET("/", indexHandler(rndr))
	router.POST("/", loginHandler(rndr))
	router.GET("/auth", authHandler)

	// middleware and static content file server
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger(),
		&negroni.Static{
			Dir:    http.Dir("public"),
			Prefix: ""})
	n.UseHandler(router)

	n.Run(fmt.Sprintf(":%d", *port))
}

func authHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var s = securecookie.New(nsCookieHashKey, nil)
	// get the cookie from the request
	if cookie, err := r.Cookie(nsCookieName); err == nil {
		value := make(map[string]string)
		// try to decode it
		if err = s.Decode(nsCookieName, cookie.Value, &value); err == nil {
			// if if succeeds set X-Forwarded-User header and return HTTP 200 status code
			w.Header().Add("X-Forwarded-User", value["user"])
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// otherwise return HTTP 401 status code
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func indexHandler(render *render.Render) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// just render the login page
		render.HTML(w, http.StatusOK, "index", nil)
	}
}

func loginHandler(render *render.Render) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		login := r.PostFormValue("login")
		passwd := r.PostFormValue("passwd")

		var errorMessage = false

		// nothing fancy here, it is just a demo so every user has the same password
		// and if it doesn't match render the login page and present user with error message
		if login == "" || passwd != cfg.Password {
			errorMessage = true
			render.HTML(w, http.StatusOK, "index", errorMessage)
		} else {
			var s = securecookie.New(nsCookieHashKey, nil)
			value := map[string]string{
				"user": login,
			}

			// encode username to secure cookie
			if encoded, err := s.Encode(nsCookieName, value); err == nil {
				cookie := &http.Cookie{
					Name:    nsCookieName,
					Value:   encoded,
					Domain:  cfg.Domain,
					Expires: time.Now().AddDate(1, 0, 0),
					Path:    "/",
				}
				http.SetCookie(w, cookie)
			}

			// after successful login redirect to original destination (if it exists)
			var redirectUrl = cfg.DefaultRedirectUrl
			if cookie, err := r.Cookie(nsRedirectCookieName); err == nil {
				redirectUrl = cookie.Value
			}
			// ... and delete the original destination holder cookie
			http.SetCookie(w, &http.Cookie{
				Name:    nsRedirectCookieName,
				Value:   "deleted",
				Domain:  cfg.Domain,
				Expires: time.Now().Add(time.Hour * -24),
				Path:    "/",
			})

			http.Redirect(w, r, redirectUrl, http.StatusFound)
		}

	}
}

// loads the config file from filename
// Example config file content:
/*
defaultRedirectUrl = "https://protected-resource.example.redbyte.eu"
password = "shared_password"
domain = ".example.redbyte.eu"
*/
func loadConfig(filename string) (config, error) {
	var cfg config
	if _, err := toml.DecodeFile(filename, &cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}

package oauth2

import (
	"context"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"net/http"
	"strconv"
	"time"
)

type OAuthToken struct {
	ClientID uint

	Token      string
	Type       string
	ExpireDate time.Time
}

func OAuthImplicitGrantAuth(url, browserPath string, clientID uint, listenPort int) (OAuthToken, error) {
	tokenC := make(chan OAuthToken)

	listenPortStr := strconv.Itoa(listenPort)
	srv := http.Server{Addr: ":" + listenPortStr}
	http.HandleFunc("/oauth2", func(w http.ResponseWriter, r *http.Request) {
		website := `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
</head>
<body>
<script>
	let params = "?" + window.location.hash.substring(1);
	let xmlHttp = new XMLHttpRequest();
    xmlHttp.open("GET", "http://localhost:` + listenPortStr + `/oauth2parsed" + params, false);
    xmlHttp.send(null);
    document.write(xmlHttp.responseText);
</script>
</body>
</html>
`
		w.Write([]byte(website))
	})
	http.HandleFunc("/oauth2parsed", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		token := OAuthToken{}
		token.ClientID = clientID
		token.Token = r.Form.Get("access_token")
		token.Type = r.Form.Get("token_type")

		expiresIn, _ := time.ParseDuration(r.Form.Get("expires_in") + "s")
		token.ExpireDate = time.Now().Add(expiresIn)

		if token.Token == "" {
			w.Write([]byte("No token received"))
			return
		}
		w.Write([]byte("Token retrieved successfully"))
		tokenC <- token
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("HTTP server error: %v\n", err))
		}
	}()

	authUrl := fmt.Sprintf("%s?client_id=%d&response_type=token", url, clientID)

	if browserPath == "" {
		open.Start(authUrl)
	} else {
		open.StartWith(authUrl, browserPath)
	}

	token := <-tokenC

	if err := srv.Shutdown(context.Background()); err != nil {
		return token, err
	}

	return token, nil
}

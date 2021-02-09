package api

import (
	"encoding/json"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/shieldwall/database"
	"io"
	"net/http"
)

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (api *API) UserRegister(w http.ResponseWriter, r *http.Request) {
	var req UserRegisterRequest

	defer r.Body.Close()

	client := clientIP(r)
	reader := io.LimitReader(r.Body, api.config.ReqMaxSize)
	decoder := json.NewDecoder(reader)

	err := decoder.Decode(&req)
	if err != nil {
		log.Warning("[%s] error decoding user registration request: %v", client, err)
		JSON(w, http.StatusBadRequest, nil)
		return
	}

	user, err := database.RegisterUser(client, req.Email, req.Password)
	if err != nil {
		log.Warning("[%s] %v", client, err)
		ERROR(w, http.StatusBadRequest, err)
		return
	}

	log.Info("[%s] registered new user %s", user.Address, user.Email)

	// prepare and send verification email
	link := fmt.Sprintf("%s/api/v1/user/verify/%s", api.config.URL, user.Verification)

	emailSubject := "shieldwall.me account verification"
	emailBody := "Follow this link to complete your registration.<br/><br/>" +
				 fmt.Sprintf("<a href=\"%s\">%s</a>", link, link)

	if err = api.sendmail.Send(api.mail.From, user.Email, emailSubject, emailBody); err != nil {
		log.Error("error sending verification email to %s: %v", user.Email, err)
	} else {
		log.Debug("verification email sent to %s (%s)", user.Email, user.Verification)
	}

	JSON(w, http.StatusOK, "registration successful, proceed to email verification")
}
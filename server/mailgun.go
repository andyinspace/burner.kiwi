package server

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// MailgunIncoming receives incoming email webhooks from mailgun. It saves the email to
// the database. Any failures return a 500. Mailgun will then retry.
func (s *Server) MailgunIncoming(w http.ResponseWriter, r *http.Request) {
	ver, err := s.mg.VerifyWebhookRequest(r)

	if err != nil {
		log.Printf("MailgunIncoming: failed to verify request: %v", err)
	}

	if !ver {
		log.Printf("MailgunIncoming: invalid request")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["inboxID"]

	i, err := s.getInboxByID(id)

	if err != nil {
		log.Printf("MailgunIncoming: failed to get inbox: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var m Message

	m.InboxID = i.ID
	m.TTL = i.TTL

	mID, err := uuid.NewRandom()

	if err != nil {
		log.Printf("MailgunIncoming: failed to generate uuid for inbox: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m.ID = mID.String()

	m.ReceivedAt = time.Now().Unix()
	m.MGID = r.FormValue("message-id")
	m.Sender = r.FormValue("sender")
	m.From = r.FormValue("from")
	m.Subject = r.FormValue("subject")
	m.BodyHTML = r.FormValue("body-html")
	m.BodyPlain = r.FormValue("body-plain")

	err = s.saveNewMessage(m)

	if err != nil {
		log.Printf("MailgunIncoming: failed to save message to db: %v", err)
	}

	_, err = w.Write([]byte(id))

	if err != nil {
		log.Printf("MailgunIncoming: failed to write response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
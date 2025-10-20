package burner

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/haydenwoodhead/burner.kiwi/stringduration"
	log "github.com/sirupsen/logrus"
)

type templateMessage struct {
	Message
	ReceivedAt   string
	AvatarLetter string
	AvatarColor  string
}

type templateInbox struct {
	Inbox
	Expires expires
}

type expires struct {
	Hours   string
	Minutes string
}

type inboxOut struct {
	Static             staticDetails
	Messages           []templateMessage
	Inbox              templateInbox
	SelectedMessage    templateMessage
	HasSelectedMessage bool
	ModalData          interface{}
}

type editModalData struct {
	Hosts []string
	Err   string
}

func transformMessagesForTemplate(msgs []Message) []templateMessage {
	transformedMsgs := make([]templateMessage, 0, len(msgs))

	// loop over all messages and calculate how long ago the message was received
	// then append that string to received to be passed to the template
	for _, m := range msgs {
		received := calculateReceivedAt(m.ReceivedAt)
		avatarLetter, avatarColor := getAvatarDetails(m.FromName)
		transformedMsgs = append(transformedMsgs, templateMessage{
			Message:      m,
			ReceivedAt:   received,
			AvatarLetter: avatarLetter,
			AvatarColor:  avatarColor,
		})
	}

	return transformedMsgs
}

func calculateReceivedAt(t int64) string {
	diff := time.Since(time.Unix(t, 0))

	// if we received the email less than 30 seconds ago then write that out
	// because rounding the duration when less than 30seconds will give us 0 seconds
	if diff.Seconds() < 30 {
		return "Less than 30s ago"
	}

	diff = diff.Round(time.Minute) // Round to nearest minute

	h, min := stringduration.GetHoursAndMinutes(diff)

	if h != "0" {
		return fmt.Sprintf("%vh %vm ago", h, min)
	}

	return fmt.Sprintf("%vm ago", min)
}

func getAvatarDetails(sender string) (string, string) {
	var letter string
	for _, runeInString := range sender {
		letter = strings.ToUpper(string(runeInString))
		break
	}

	// TOODO: normalize unicode so we get a better distribution of colors
	if letter < "E" {
		return letter, "bg-red"
	} else if letter >= "E" && letter < "I" {
		return letter, "bg-yellow"
	} else if letter >= "I" && letter < "M" {
		return letter, "bg-green"
	} else if letter >= "M" && letter < "Q" {
		return letter, "bg-indigo"
	} else if letter >= "Q" && letter < "U" {
		return letter, "bg-purple"
	}

	return letter, "bg-pink"
}

func transformInboxForTemplate(i Inbox) templateInbox {
	expiration := time.Until(time.Unix(i.TTL, 0))
	h, m := stringduration.GetHoursAndMinutes(expiration)

	return templateInbox{
		Inbox: i,
		Expires: expires{
			Hours:   h,
			Minutes: m,
		},
	}
}

//go:embed templates
var embeddedTemplatesDir embed.FS

// Templates
var indexTemplate *template.Template
var indexTemplateOnce *sync.Once = &sync.Once{}

var editTemplate *template.Template
var editTemplateOnce sync.Once

var deleteTemplate *template.Template
var deleteTemplateOnce sync.Once

var extendTemplate *template.Template
var extendTemplateOnce sync.Once

func (s *Server) parseTemplate(name string, parts ...string) (*template.Template, error) {
	t := template.New(name)

	var templatesDir fs.FS
	if s.cfg.Developing {
		templatesDir = os.DirFS("./burner/templates")
	} else {
		subFs, err := fs.Sub(embeddedTemplatesDir, "templates")
		if err != nil {
			log.WithField("dev", s.cfg.Developing).WithError(err).Fatal("getTemplate: failed to get sub fs")
			return nil, nil
		}
		templatesDir = subFs
	}

	return t.ParseFS(templatesDir, parts...)
}

func (s *Server) getIndexTemplate() *template.Template {
	gen := func() *template.Template {
		t, err := s.parseTemplate("index", "base.html", "inbox.html", "emptyModal.html")
		if err != nil {
			log.WithError(err).Fatal("getIndexTemplate: failed to get")
			return nil
		}
		return t
	}

	if s.cfg.Developing {
		t := gen()
		return t
	}

	indexTemplateOnce.Do(func() {
		t := gen()
		indexTemplate = t
	})

	return indexTemplate
}

func (s *Server) getEditTemplate() *template.Template {
	gen := func() *template.Template {
		t, err := s.parseTemplate("index", "base.html", "inbox.html", "edit.html")
		if err != nil {
			log.WithError(err).Fatal("getEditTemplate: failed to get")
			return nil
		}
		return t
	}

	if s.cfg.Developing {
		t := gen()
		return t
	}

	editTemplateOnce.Do(func() {
		t := gen()
		editTemplate = t
	})

	return editTemplate
}

func (s *Server) getDeleteTemplate() *template.Template {
	gen := func() *template.Template {
		t, err := s.parseTemplate("index", "base.html", "inbox.html", "delete.html")
		if err != nil {
			log.WithError(err).Fatal("getDeleteTemplate: failed to get")
			return nil
		}
		return t
	}

	if s.cfg.Developing {
		t := gen()
		return t
	}

	deleteTemplateOnce.Do(func() {
		t := gen()
		deleteTemplate = t
	})

	return deleteTemplate
}

func (s *Server) getExtendTemplate() *template.Template {
	gen := func() *template.Template {
		t, err := s.parseTemplate("index", "base.html", "inbox.html", "extend.html")
		if err != nil {
			log.WithError(err).Fatal("getExtendTemplate: failed to get")
			return nil
		}
		return t
	}

	if s.cfg.Developing {
		t := gen()
		return t
	}

	extendTemplateOnce.Do(func() {
		t := gen()
		extendTemplate = t
	})

	return extendTemplate
}

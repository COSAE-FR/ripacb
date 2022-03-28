package entity

import (
	"fmt"
	"strings"
	"time"
)

const revisionTemplate = `%s++++
%s
`

const (
	defaultUser = "firewall"
)

type Revision struct {
	Revision   string `validate:"required,min=6"`
	Hash       string `validate:"hexadecimal,required,len=64"`
	Content    string `validate:"required"`
	Reason     string `validate:"required"`
	Device     string `validate:"hexadecimal,required,len=64"`
	Username   string
	Comment    string
	Date       time.Time
	FromPortal bool
}

func (r *Revision) MarshallText() string {
	return fmt.Sprintf(revisionTemplate, r.Hash, r.Content)
}

func (r *Revision) Label() string {
	if r.Comment != "" {
		return fmt.Sprintf("%s: %s", r.Username, r.Comment)
	}
	return r.Reason
}

type RevisionList map[string]Revision

func (l RevisionList) MarshallText() string {
	var ret []string
	for _, rev := range l {
		ret = append(ret, strings.Join([]string{rev.Username, rev.Reason, rev.Date.UTC().Format(time.RFC3339)}, "||"))
	}
	return strings.Join(ret, "\n")
}

func ParseReason(reason string) (string, string) {
	user := defaultUser
	comment := ""
	reasonParts := strings.Split(reason, ":")
	if len(reasonParts) > 1 {
		comment = strings.TrimSpace(strings.Join(reasonParts[1:], ":"))
		reasonParts = strings.Split(reasonParts[0], "@")
		user = reasonParts[0]
	}
	return user, comment
}

package restore

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/cliconfig"
	"github.com/COSAE-FR/riputils/pfsense/configuration/sections"
	"github.com/google/uuid"
	"github.com/rivo/tview"
	"io"
	"strings"
)

func errorModal(pages *tview.Pages, nextPage string, text string, args ...interface{}) {
	errModalId := uuid.NewString()
	errModal := tview.NewModal()
	errModal.SetText(fmt.Sprintf(text, args...))
	errModal.AddButtons([]string{"OK"})
	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.SwitchToPage(nextPage)
		pages.RemovePage(errModalId)
	})
	pages.AddPage(errModalId, errModal, false, true)
}

func patchXml(original []byte) ([]byte, error) {
	inBuf := bytes.NewBuffer(original)

	var outBuf bytes.Buffer
	decoder := xml.NewDecoder(inBuf)
	encoder := xml.NewEncoder(&outBuf)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return original, err
		}

		switch v := token.(type) {
		case xml.StartElement:
			if v.Name.Local == "acb" {
				var desc sections.ACB
				if err = decoder.DecodeElement(&desc, &v); err != nil {
					return original, err
				}
				desc.Password = cliconfig.Config.Password
				desc.Server = cliconfig.Config.ServerURL
				if err = encoder.EncodeElement(desc, v); err != nil {
					return original, err
				}
				continue
			}
		}

		if err := encoder.EncodeToken(xml.CopyToken(token)); err != nil {
			return original, err
		}
	}

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		return original, err
	}
	return outBuf.Bytes(), nil
}

func deviceKey(hostname, password string) string {
	h := hmac.New(sha256.New, []byte(password))
	h.Write([]byte(strings.TrimSpace(strings.ToLower(hostname))))
	dk := h.Sum(nil)
	return hex.EncodeToString(dk)
}

package restore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/cliconfig"
	"github.com/COSAE-FR/ripacb/pkg/acb/bindings"
	"github.com/COSAE-FR/ripacb/pkg/acb/entity"
	"github.com/COSAE-FR/riputils/common"
	"github.com/Luzifer/go-openssl/v4"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func restoreBackup(id string, progress chan int) error {
	progress <- 1
	rev, err := time.Parse(time.RFC3339, id)
	if err != nil {
		return err
	}
	id = rev.UTC().Format(time.RFC3339)
	progress <- 1
	dk := deviceKey(cliconfig.Config.Hostname, cliconfig.Config.Password)
	progress <- 1
	req := bindings.GetBackupRequest{
		Version:   "22.2",
		DeviceKey: dk,
		Revision:  id,
	}
	progress <- 1
	body, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	progress <- 1
	client := http.Client{}
	post, err := client.Post(cliconfig.Config.ServerURL+"/api/v1/backups", "application/json", bytes.NewBuffer(body))
	progress <- 1
	if err != nil {
		return err
	}
	if post.Body != nil {
		defer post.Body.Close()
	}
	if post.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid server response: %s (%d)\nfor revision %s", post.Status, post.StatusCode, id)
	}
	progress <- 1
	var revision entity.Revision
	raw, err := ioutil.ReadAll(post.Body)
	if err != nil {
		return err
	}
	progress <- 1
	err = json.Unmarshal(raw, &revision)
	if err != nil {
		return err
	}
	progress <- 1
	pass := openssl.NewPBKDF2Generator(sha256.New, 10000)
	o := openssl.New()
	dec, err := o.DecryptBytes(cliconfig.Config.Password, prepareEncryptedText(revision.Content), pass)
	if err != nil {
		return err
	}
	progress <- 1
	decHash := sha256.Sum256(dec)
	if revision.Hash != hex.EncodeToString(decHash[:]) {
		return fmt.Errorf("decrypted data is corrupted")
	}
	progress <- 1
	conf := string(dec)
	if !strings.Contains(conf, "<pfsense>") {
		return fmt.Errorf("cleartext data seems invalid")
	}
	progress <- 1
	input, err := ioutil.ReadFile(cliconfig.PfSenseXML)
	if err != nil {
		return fmt.Errorf("cannot read old configuration file at %s", cliconfig.PfSenseXML)
	}
	progress <- 1
	patched, err := patchXml(dec)
	if err == nil {
		dec = patched
	}
	err = ioutil.WriteFile(cliconfig.PfSenseXML, dec, 0644)
	if err != nil {
		_ = ioutil.WriteFile(cliconfig.PfSenseXML, input, 0644)
		return err
	}
	if common.FileExists("/conf/trigger_initial_wizard") {
		_ = os.Remove("/conf/trigger_initial_wizard")
	}

	return nil

}

func prepareEncryptedText(original string) []byte {
	original = strings.TrimSpace(original)
	b64 := regexp.MustCompile("----\\s+BEGIN\\s+config\\.xml\\s+-+\\n+([^\\-]+)\\n+-+\\s+END\\s+config\\.xml\\s+-+")
	results := b64.FindStringSubmatch(original)
	if len(results) == 2 {
		return []byte(results[1])
	}
	return []byte(original)
}

package config

import (
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"strings"

	"github.com/coreos/coreos-cloudinit/pkg"
)

type MimeMultiPart struct {
	Scripts []*Script
	Configs []*CloudConfig
}

func (m *MimeMultiPart) AddScript(script []byte) error {
	s, err := NewScript(string(script))
	if err == nil {
		m.Scripts = append(m.Scripts, s)
	}
	return err
}

func (m *MimeMultiPart) AddCloudConfig(config []byte) error {
	c, err := NewCloudConfig(string(config))
	if err == nil {
		m.Configs = append(m.Configs, c)
	}
	return err
}

func (m *MimeMultiPart) AddMimeMultiPart(mmp []byte) error {
	mm, err := NewMimeMultiPart(string(mmp))
	if err == nil {
		m.Config.Merge(mm.Config)
		for _, script := range mm.Scripts {
			if !m.scriptAlreadyExists(script) {
				m.Scripts = append(m.Scripts, script)
			}
		}
	}
	return err
}

func (m *MimeMultiPart) scriptAlreadyExists(script *Script) bool {
	for _, s := range m.Scripts {
		if s == script {
			return true
		}
	}
	return false
}

func (m *MimeMultiPart) AddPlainText(pt []byte) error {
	switch {
	case IsScript(string(pt)):
		return m.AddScript(pt)
	case IsCloudConfig(string(pt)):
		return m.AddCloudConfig(pt)
	case IsMimeMultiPart(string(pt)):
		return m.AddMimeMultiPart(pt)
	}
	return nil
}

func (m *MimeMultiPart) fileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func (m *MimeMultiPart) readUrl(url string) ([]byte, error) {
	client := pkg.NewHttpClient()
	return client.GetRetry(url)
}

func (m *MimeMultiPart) AddUrl(url string) error {
	var err error
	if data, err := m.readUrl(url); err == nil {
		err = m.AddPlainText(data)
	}
	return err
}

func (m *MimeMultiPart) AddUrls(urls []byte) error {
	var err error
	for _, url := range strings.Split(string(urls), "\n") {
		err = m.AddUrl(url)
		if err != nil {
			break
		}
	}
	return err
}

func IsMimeMultiPart(userdata string) bool {
	r := strings.NewReader(userdata)
	m, err := mail.ReadMessage(r)
	if err != nil {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil {
		return false
	}
	return strings.HasPrefix(mediaType, "multipart/")
}

func NewMimeMultiPart(content string) (*MimeMultiPart, error) {
	var mmp MimeMultiPart

	r := strings.NewReader(content)
	m, err := mail.ReadMessage(r)
	if err != nil {
		return nil, err
	}
	mediaType, params, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(m.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			slurp, err := ioutil.ReadAll(p)
			if err != nil {
				return nil, err
			}
			switch p.Header.Get("Content-Type") {
			case "text/x-shellscript":
				err = mmp.AddScript(slurp)
			case "text/cloud-config":
				err = mmp.AddCloudConfig(slurp)
			case "text/plain":
				err = mmp.AddPlainText(slurp)
			case "text/x-include-url":
				err = mmp.AddUrls(slurp)
			case "text/x-include-once-url":
				err = mmp.AddUrls(slurp)
			}
			if err != nil {
				return nil, err
			}
		}
	}
	return &mmp, nil
}

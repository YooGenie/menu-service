package common

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"html/template"
	"mime/multipart"
	"net/smtp"
	"path/filepath"
	"strings"
	"study-service/config"
)

type Message struct {
	To          []string
	CC          []string
	BCC         []string
	autoBcc     bool
	Body        string
	Subject     string
	Sender      string
	Attachments map[string][]byte
}

func NewMessage(sender string, subject string, autoBcc bool) *Message {
	return &Message{Subject: subject, Sender: sender, Attachments: make(map[string][]byte)}
}

// 첨부파일
func (m *Message) AttachTemplate(filePath string, data interface{}) error {
	_, fileName := filepath.Split(filePath)

	// parse html
	strTemplate, err := ParseHtmlTemplate(filePath, data)
	if err != nil {
		return err
	}
	// fmt.Println(strTemplate)

	// escape html
	escapedString := template.JSEscaper(strTemplate)
	htmlStr := fmt.Sprintf(`<script language=javascript>document.write(unescape("%s"));</script>`, escapedString)

	m.Attachments[fileName] = []byte(htmlStr)

	return nil
}

// 바디 셋팅하기
func (m *Message) SetMailBody(filePath string, data interface{}) error {
	// parse html
	strTemplate, err := ParseHtmlTemplate(filePath, data)
	if err != nil {
		return err
	}

	m.Body = strTemplate

	return nil
}

func (m *Message) Send() error {
	return m.send()
}

func (m *Message) send() error {
	// connect
	client, err := m.connect()
	if err != nil {
		fmt.Println("에러 : ", err)
		return err
	}
	defer client.Close()

	// set sender & receiver
	if err := client.Mail(m.Sender); err != nil {
		return errors.Wrap(err, "")
	}

	receivers := m.To
	if len(m.CC) > 0 {
		receivers = append(receivers, m.CC...)
	}
	if len(m.BCC) > 0 {
		receivers = append(receivers, m.BCC...)
	}
	if m.autoBcc {
		receivers = append(receivers, []string{m.Sender}...)
	}

	for _, to := range receivers {
		if err := client.Rcpt(to); err != nil { //RCPT TO (Recipient To) 받는 사람의 전자 메일 주소를 지정
			return errors.Wrap(err, "")
		}
	}

	// write body
	writer, err := client.Data()
	if err != nil {
		return errors.Wrap(err, "")
	}

	if _, err := writer.Write(m.ToBytes()); err != nil {
		return errors.Wrap(err, "")
	}
	if err := writer.Close(); err != nil {
		return errors.Wrap(err, "")
	}

	client.Quit()

	return err
}

// smtp로 연결하는 곳
func (m *Message) connect() (*smtp.Client, error) {
	//TLS : 개인 정보 보호 및 안전한 전송을 위해 이메일을 암호화하는 표준 인터넷 프로토콜

	tlsConfig := tls.Config{
		ServerName:         config.Config.Mail.Host,
		InsecureSkipVerify: true,
	}

	// 지정된 네트워크 주소에 연결
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", config.Config.Mail.Host, config.Config.Mail.Port), &tlsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	// defer conn.Close()

	//인증 시 사용할 서버 이름으로 기존 연결 및 호스트를 사용하는 새 클라이언트를 반환한다.
	client, clientErr := smtp.NewClient(conn, config.Config.Mail.Host)
	if clientErr != nil {
		return nil, errors.Wrap(err, "")
	}
	// defer client.Close()

	// 사용자이름과 암호 사용하여 호스트를 인증
	auth := smtp.PlainAuth("", config.Config.Mail.User, config.Config.Mail.Password, config.Config.Mail.Host)
	if err := client.Auth(auth); err != nil { //클라이언트 인증
		return nil, errors.Wrap(err, "")
	}

	return client, nil
}

// 메일 제목, 메일 내용을 바이트로 바꾸는 것
func (m *Message) ToBytes() []byte {
	withAttachments := len(m.Attachments) > 0

	buf := bytes.NewBuffer(nil)

	buf.WriteString(fmt.Sprintf("Subject: =?UTF-8?B?%s?=\r\n", base64.StdEncoding.EncodeToString([]byte(m.Subject))))
	buf.WriteString(fmt.Sprintf("From: %s\r\n", m.Sender))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(m.To, ",")))
	if len(m.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(m.CC, ",")))
	}
	if len(m.BCC) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(m.BCC, ",")))
	}

	buf.WriteString("MIME-Version: 1.0\r\n")
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))

	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")

	buf.WriteString(fmt.Sprintf("\r\n%s", m.Body))
	buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))

	if withAttachments {
		for k, v := range m.Attachments {
			buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: Application/octet-stream; name=%s\r\n", k))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment;\r\n\r\n"))

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)
			buf.Write(b)
			buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))
		}

		buf.WriteString("--")
	}

	return buf.Bytes()
}

package verifier

import (
	"bufio"
	"fmt"
	"log/slog"
	"mailnull/api/internal/config"
	"math/rand"
	"net"
	"net/textproto"
	"regexp"
	"strings"
	"time"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	disposableDomains = map[string]bool{
		"mailinator.com":    true,
		"guerrillamail.com": true,
		"temp-mail.org":     true,
	}
)

type Verifier struct {
	cfg *config.Config
	log *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) *Verifier {
	return &Verifier{
		cfg: cfg,
		log: log,
	}
}

func (v *Verifier) Verify(email string) Result {
	res := Result{
		Email:     email,
		Timestamp: time.Now(),
		Provider:  "unknown",
	}

	if !emailRegex.MatchString(email) {
		res.Deliverability = Undeliverable
		res.QualityScore = 0.0
		return res
	}
	res.IsValidFormat = true

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		res.Deliverability = Undeliverable
		return res
	}
	domain := parts[1]
	res.Provider = domain

	if disposableDomains[domain] {
		res.IsDisposableEmail = true
		res.Deliverability = Risky
		res.QualityScore = 0.4
		return res
	}

	if v.cfg.Mode == "MOCK" {
		return v.mockVerify(res)
	}

	mxRecords, err := net.LookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		res.Deliverability = Undeliverable
		res.QualityScore = 0.0
		return res
	}

	if v.cfg.Mode == "LITE" {
		res.Deliverability = Deliverable
		res.QualityScore = 0.7
		return res
	}

	deliverable, isCatchAll := v.checkSMTP(domain, mxRecords, email)

	if deliverable && isCatchAll {
		res.Deliverability = Risky
		res.QualityScore = 0.5
		res.Error = "Catch-all domain detected"
	} else if deliverable {
		res.Deliverability = Deliverable
		res.QualityScore = 0.9
	} else if isCatchAll {
		res.Deliverability = Risky
		res.QualityScore = 0.5
		res.Error = "Verification failed due to network/timeout"
	} else {
		res.Deliverability = Undeliverable
		res.QualityScore = 0.1
	}

	return res
}

func (v *Verifier) checkSMTP(domain string, mxRecords []*net.MX, email string) (bool, bool) {
	target := fmt.Sprintf("%s:25", mxRecords[0].Host)

	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		v.log.Info("SMTP Dial failed", "domain", domain, "error", err)
		return false, true
	}
	defer conn.Close()

	client := newSMTPClient(conn)

	if _, err := client.ReadCodeLine(220); err != nil {
		return false, true
	}

	if err := client.SendCommand("EHLO mailnull.com"); err != nil {
		if err := client.SendCommand("HELO mailnull.com"); err != nil {
			return false, true
		}
	}

	if err := client.SendCommand("MAIL FROM:<verify@mailnull.com>"); err != nil {
		return false, true
	}

	code, err := client.SendRecpt("RCPT TO:<" + email + ">")
	if err != nil {
		return false, true
	}

	if code == 250 {

		if v.isCatchAll(client, domain) {
			return true, true
		}
		return true, false
	}

	if code >= 500 && code < 600 {
		return false, false
	}

	return false, true
}

func (v *Verifier) isCatchAll(client *SMTPClient, domain string) bool {
	randomUser := fmt.Sprintf("chk_%d", rand.Intn(100000))
	code, err := client.SendRecpt("RCPT TO:<" + randomUser + "@" + domain + ">")
	if err != nil {
		return false
	}
	return code == 250
}

func (v *Verifier) mockVerify(res Result) Result {

	h := 0
	for _, c := range res.Email {
		h += int(c)
	}

	r := h % 100
	if r < 10 {
		res.Deliverability = Undeliverable
		res.QualityScore = 0.1
	} else if r < 30 {
		res.Deliverability = Risky
		res.QualityScore = 0.5
	} else {
		res.Deliverability = Deliverable
		res.QualityScore = 0.95
	}
	return res
}

type SMTPClient struct {
	conn   net.Conn
	reader *textproto.Reader
}

func newSMTPClient(conn net.Conn) *SMTPClient {
	return &SMTPClient{
		conn:   conn,
		reader: textproto.NewReader(bufio.NewReader(conn)),
	}
}

func (c *SMTPClient) ReadCodeLine(expectCode int) (string, error) {
	line, err := c.reader.ReadLine()
	if err != nil {
		return "", err
	}
	if len(line) < 3 {
		return line, fmt.Errorf("short line")
	}
	return line, nil
}

func (c *SMTPClient) SendCommand(cmd string) error {
	_, err := fmt.Fprintf(c.conn, "%s\r\n", cmd)
	if err != nil {
		return err
	}
	_, err = c.reader.ReadLine()
	return err
}

func (c *SMTPClient) SendRecpt(cmd string) (int, error) {
	_, err := fmt.Fprintf(c.conn, "%s\r\n", cmd)
	if err != nil {
		return 0, err
	}
	line, err := c.reader.ReadLine()
	if err != nil {
		return 0, err
	}
	var code int
	fmt.Sscanf(line, "%d", &code)
	return code, nil
}

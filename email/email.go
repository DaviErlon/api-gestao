package email

import (
    "fmt"
    "net/smtp"
    "os"
)

// Config armazena as configurações SMTP
type config struct {
    Host     string
    Port     string
    User     string
    Password string
    From     string
}

var EMAIL config

// LoadConfig carrega das variáveis de ambiente
func init() {
    EMAIL = config{
        Host:     os.Getenv("SMTP_HOST"),
        Port:     os.Getenv("SMTP_PORT"),
        User:     os.Getenv("SMTP_USER"),
        Password: os.Getenv("SMTP_PASS"),
        From:     os.Getenv("SMTP_FROM"),
    }
}

func (c *config) Send(to, subject, bodyHTML string) error {
    // Autenticação
    auth := smtp.PlainAuth("", c.User, c.Password, c.Host)

    // Monta as headers
    headers := make(map[string]string)
    headers["From"] = c.From
    headers["To"] = to
    headers["Subject"] = subject
    headers["MIME-Version"] = "1.0"
    headers["Content-Type"] = "text/html; charset=utf-8"

    // Constrói a mensagem
    msg := ""
    for k, v := range headers {
        msg += fmt.Sprintf("%s: %s\r\n", k, v)
    }
    msg += "\r\n" + bodyHTML

    // Endereço do servidor
    addr := c.Host + ":" + c.Port
    return smtp.SendMail(addr, auth, c.User, []string{to}, []byte(msg))
}
package external

import (
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	gomail "gopkg.in/gomail.v2"
)

// EmailService отправляет email-уведомления через SMTP.
type EmailService struct {
	host     string
	port     int
	user     string
	password string
	from     string
	logger   *logrus.Logger
}

func NewEmailService(host, portStr, user, password, from string, logger *logrus.Logger) *EmailService {
	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 587
	}
	return &EmailService{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
		logger:   logger,
	}
}

// SendPaymentNotification уведомляет клиента о списании платежа по кредиту.
func (s *EmailService) SendPaymentNotification(toEmail, userName string, amount float64, creditID string) {
	if s.user == "" || s.password == "" {
		s.logger.Warn("SMTP not configured, skipping email notification")
		return
	}

	subject := "Уведомление о платеже по кредиту"
	body := fmt.Sprintf(
		"Уважаемый %s,\n\nСо счёта списан платёж по кредиту %s на сумму %.2f RUB.\n\nС уважением,\nБанковский сервис",
		userName, creditID, amount,
	)
	s.send(toEmail, subject, body)
}

// SendOverdueNotification уведомляет о просрочке и начислении штрафа.
func (s *EmailService) SendOverdueNotification(toEmail, userName string, amount, fineAmount float64, creditID string) {
	if s.user == "" || s.password == "" {
		s.logger.Warn("SMTP not configured, skipping overdue email")
		return
	}

	subject := "Просрочка платежа по кредиту"
	body := fmt.Sprintf(
		"Уважаемый %s,\n\nПлатёж по кредиту %s на сумму %.2f RUB просрочен.\n"+
			"Начислен штраф: %.2f RUB.\n\nПожалуйста, пополните счёт.\n\nС уважением,\nБанковский сервис",
		userName, creditID, amount, fineAmount,
	)
	s.send(toEmail, subject, body)
}

// SendRegistrationWelcome отправляет приветственное письмо новому пользователю.
func (s *EmailService) SendRegistrationWelcome(toEmail, userName string) {
	if s.user == "" || s.password == "" {
		return
	}
	subject := "Добро пожаловать в Банковский сервис"
	body := fmt.Sprintf(
		"Уважаемый %s,\n\nВаш аккаунт успешно создан.\n\nС уважением,\nБанковский сервис",
		userName,
	)
	s.send(toEmail, subject, body)
}

func (s *EmailService) send(to, subject, body string) {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain; charset=utf-8", body)

	d := gomail.NewDialer(s.host, s.port, s.user, s.password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		s.logger.WithError(err).Errorf("Failed to send email to %s", to)
	} else {
		s.logger.Infof("Email sent to %s: %s", to, subject)
	}
}

package auth

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailService handles sending emails
type EmailService struct {
	host     string
	port     int
	email    string
	password string
	auth     smtp.Auth
}

// NewEmailService creates a new email service
func NewEmailService(host string, port int, email, password string) *EmailService {
	auth := smtp.PlainAuth("", email, password, host)
	return &EmailService{
		host:     host,
		port:     port,
		email:    email,
		password: password,
		auth:     auth,
	}
}

// SendOTP sends an OTP to the specified email address
func (e *EmailService) SendOTP(to, otp string) error {
	subject := "Your Chat Server OTP Code"
	body := e.formatOTPEmail(otp)

	message := e.formatEmail(e.email, to, subject, body)

	addr := fmt.Sprintf("%s:%d", e.host, e.port)
	err := smtp.SendMail(addr, e.auth, e.email, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// formatEmail formats an email message
func (e *EmailService) formatEmail(from, to, subject, body string) string {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	return message.String()
}

// formatOTPEmail creates a formatted HTML email for OTP
func (e *EmailService) formatOTPEmail(otp string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #4CAF50;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border-radius: 0 0 5px 5px;
        }
        .otp-code {
            background-color: #fff;
            border: 2px solid #4CAF50;
            border-radius: 5px;
            padding: 20px;
            text-align: center;
            font-size: 32px;
            font-weight: bold;
            letter-spacing: 5px;
            margin: 20px 0;
            color: #4CAF50;
        }
        .footer {
            margin-top: 20px;
            font-size: 12px;
            color: #666;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>TCP Chat Server</h1>
        </div>
        <div class="content">
            <h2>Your One-Time Password</h2>
            <p>You have requested to authenticate with our chat server. Please use the following OTP code to complete your login:</p>
            
            <div class="otp-code">%s</div>
            
            <p><strong>Important:</strong></p>
            <ul>
                <li>This code will expire in 5 minutes</li>
                <li>This code can only be used once</li>
                <li>Do not share this code with anyone</li>
            </ul>
            
            <p>If you did not request this code, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>This is an automated message from TCP Chat Server. Please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>
`, otp)
}

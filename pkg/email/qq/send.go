package qq

import (
	"bytes"
	"fmt"
	sendEmail "github.com/jordan-wright/email"
	"github.com/spf13/viper"
	"html/template"
	"net/smtp"
)

// SendEmail 使用第三方库发送邮件
func SendEmail(Subject, fromUser, toUser, Code string) error {
	e := sendEmail.NewEmail()

	e.From = fromUser
	e.To = []string{toUser}
	e.Subject = Subject

	t, err := template.ParseFiles("pkg/email/qq/send.html")
	if err != nil {
		return err
	}
	body := new(bytes.Buffer)
	//作为变量传递给html模板
	err1 := t.Execute(body, struct {
		Email      string
		ActiveCode string
	}{
		Email:      toUser,
		ActiveCode: Code,
	})
	if err1 != nil {
		return err1
	}
	// html形式的消息
	e.HTML = body.Bytes()
	fmt.Println(viper.GetString("email.qq_email_addr"), smtp.PlainAuth(
		"",
		viper.GetString("email.qq_email_user"),
		viper.GetString("email.qq_email_password"),
		viper.GetString("email.qq_email_host"),
	))
	return e.Send(viper.GetString("email.qq_email_addr"), smtp.PlainAuth(
		"",
		viper.GetString("email.qq_email_user"),
		viper.GetString("email.qq_email_password"),
		viper.GetString("email.qq_email_host"),
	))
}

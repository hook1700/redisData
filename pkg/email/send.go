package email

import (
	"github.com/spf13/viper"
	"redisData/pkg/email/google"
	"redisData/pkg/email/qq"
)

func SendEmail(title string, text string, toId string) error {
	var err error
	emailType := viper.GetString("email.email_type")
	if emailType == "QQ" {
		err = qq.SendEmail(title, viper.GetString("email.qq_email_user"), toId, text)
		if err != nil {
			return err
		}
	}
	if emailType == "GOOGLE" {
		err = google.New().Send(title, text, toId)
		if err != nil {
			return err
		}
	}
	return nil
}

package gus

import (
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"os"
)

type Mailer struct {
	ss *ses.SES
}

func NewSesMailer() *Mailer {
	os.Setenv("AWS_REGION", "us-west-2")
	sess := session.Must(session.NewSession())
	ss := ses.New(sess)
	return &Mailer{ss: ss}
}

type MessageParams struct {
	Subject string
	Message string
	FromEmail string
	ToEmail string
}

func (m *Mailer) Send(p MessageParams) error {
	Debug("DISABLED EMAIL SENDING:", p)
	return nil
	params := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(p.ToEmail), // Required
			},
		},
		Message: &ses.Message{ // Required
			Body: &ses.Body{ // Required
				//Html: &ses.Content{
				//	Data:    aws.String("MessageData"), // Required
				//	Charset: aws.String("Charset"),
				//},
				Text: &ses.Content{
					Data:    aws.String(p.Message), // Required
				},
			},
			Subject: &ses.Content{ // Required
				Data:    aws.String(p.Subject), // Required
			},
		},
		Source:               aws.String(p.FromEmail), // Required
		//ConfigurationSetName: aws.String("ConfigurationSetName"),
		ReplyToAddresses: []*string{
			aws.String(p.FromEmail), // Required
		},
		ReturnPath:    aws.String(p.FromEmail),
		//ReturnPathArn: aws.String("AmazonResourceName"),
		//SourceArn:     aws.String("AmazonResourceName"),
		Tags: []*ses.MessageTag{
			{ // Required
				Name:  aws.String("MessageCategory"),  // Required
				Value: aws.String("Transactional"), // Required
			},
		},
	}
	resp, err := m.ss.SendEmail(params)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		LogErr(err)
		return err
	}
	Debug(fmt.Sprintf("MAIL RES: %+v", resp))
	return nil
}

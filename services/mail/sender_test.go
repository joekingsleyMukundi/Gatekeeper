package mail

import (
	"testing"

	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	config, err := utils.LoadConfig("../../")
	require.NoError(t, err)
	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	<p>This is a test message from <a href="http://techschool.guru">Tech School</a></p>
	`
	to := []string{"joekingsleymukundi@gmail.com"}

	err = sender.SendEmail(subject, content, to, nil, nil, nil)
	require.NoError(t, err)
}

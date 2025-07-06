package webhooks

import "github.com/sirupsen/logrus"

type WebHooks interface {
	SendMessageAboutAnotherIp()
}

type WebHooksClient struct {
}

func NewClient() *WebHooksClient {
	return &WebHooksClient{}
}

func (wh *WebHooksClient) SendMessageAboutAnotherIp() {
	logrus.Info("message has been sent")
}

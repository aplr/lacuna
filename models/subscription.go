package models

import "strings"

type Subscription struct {
	Service  string
	Name     string
	Topic    string
	Endpoint string
	Deadline uint16
}

func (s *Subscription) GetSubscriptionID() string {
	return strings.Join([]string{s.Service, s.Name}, "_")
}

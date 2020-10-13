package notification

type MockNotifier struct {
	Notifications []Notification
}

type Notification struct {
	channel string
	message string
}

var _ Notifier = (*MockNotifier)(nil)

func NewMock() *MockNotifier {
	return &MockNotifier{
		Notifications: []Notification{},
	}
}

func (m *MockNotifier) Notify(channel string, message string) error {
	m.Notifications = append(m.Notifications, Notification{channel, message})
	return nil
}

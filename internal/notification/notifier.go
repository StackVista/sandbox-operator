package notification

type Notifier interface {
	Notify(channel string, message string) error
}

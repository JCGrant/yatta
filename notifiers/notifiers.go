package notifiers

// Notifier will notify a user with a message
type Notifier interface {
	Notify(message string) error
}

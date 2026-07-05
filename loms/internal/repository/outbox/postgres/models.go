package postgres

type Kind string

const (
	KindNotification Kind = "notification"
)

type Message struct {
	IdempotencyKey string
	Kind           Kind
	Payload        []byte
}

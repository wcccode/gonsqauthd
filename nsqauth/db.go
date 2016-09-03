package nsqauth

type Entry struct {
	Login       string
	Ip          string
	TlsRequired bool
	Topic       string
	Channel     string
	Subscribe   string
	Publish     string
}

type DB interface {
	Get(login, ip string, tlsRequired bool) []Entry
	List() []Entry
}

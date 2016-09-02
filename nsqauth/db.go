package nsqauth

type DB interface {
	Get(login, ip, tlsRequired string) []string
	List() [][]string
}

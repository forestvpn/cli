package wireguard

type Interface struct {
	Address    string
	PrivateKey string
	DNS        string
}

type Peer struct {
	PublicKey           string
	Endpoint            string
	AllowedIPs          []string
	PersistentKeepalive int
}

type WireguardClientConfig struct {
	Interface
	Peer
}

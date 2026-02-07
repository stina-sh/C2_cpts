package beacon

type Beaconer interface {
	Beacon() (string, string)
}


func NewHttpAuthBeacon(sysid, url, agent string, typ string) *HttpAuthBeacon {
	h := new(HttpAuthBeacon)

	h.id = sysid
	h.agent = agent
	h.url = url
	h.typ = typ

	return h
}

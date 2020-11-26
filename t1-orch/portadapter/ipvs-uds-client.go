package portadapter

type IpvsEntity struct {
}

func NewIpvsEntity() *IpvsEntity {
	return &IpvsEntity{}
}

func (ipvsEntity *IpvsEntity) NewIPVSService(string, uint16, uint32, string, uint16, string) error {
	return nil
}

func (ipvsEntity *IpvsEntity) AddIPVSApplicationServersForService(string, uint16, uint32, string, uint16, map[string]uint16, string) error {
	return nil
}

func (ipvsEntity *IpvsEntity) RemoveIPVSService(string, uint16, uint16, string) error {
	return nil
}

func (ipvsEntity *IpvsEntity) RemoveIPVSApplicationServersFromService(string, uint16, uint32, string, uint16, map[string]uint16, string) error {
	return nil
}

func (ipvsEntity *IpvsEntity) GetIPVSRuntime(string) (map[string]map[string]uint16, error) {
	return nil, nil
}

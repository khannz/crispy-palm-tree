package portadapter

import (
	"fmt"
	"strings"

	"github.com/khannz/crispy-palm-tree/domain"
)

const (
	rawCommandsForTCPHealthchecks = `ifconfig lo:10 SERVICE_IP netmask 255.255.255.255 -arp up
	ifconfig tunl0 up
	sysctl -w net.ipv4.conf.tunl0.rp_filter=0
	sysctl -w net.ipv4.conf.all.rp_filter=0
	
	net.ipv4.ip_forward = 0
	
	iptables -t nat -A PREROUTING -i tunl0 -p tcp -d APPLICATION_SERVER_IP --dport APPLICATION_SERVER_HEALTHCHECK_PORT -j DNAT --to-destination SERVICE_IP:SERVICE_PORT
`

	rawCommandsForICMPHealthchecks = `ifconfig lo:10 SERVICE_IP netmask 255.255.255.255 -arp up
ifconfig tunl0 up
sysctl -w net.ipv4.conf.tunl0.rp_filter=0
sysctl -w net.ipv4.conf.all.rp_filter=0

net.ipv4.ip_forward = 0

iptables -t nat -A PREROUTING -i tunl0 -p tcp -d APPLICATION_SERVER_IP -j DNAT --to-destination SERVICE_IP:SERVICE_PORT
`

	rawCommandsForHTTPHealthchecks = `ifconfig lo:10 SERVICE_IP netmask 255.255.255.255 -arp up
ifconfig tunl0 up
sysctl -w net.ipv4.conf.tunl0.rp_filter=0
sysctl -w net.ipv4.conf.all.rp_filter=0

net.ipv4.ip_forward = 0

iptables -t nat -A PREROUTING -i tunl0 -p tcp -d APPLICATION_SERVER_IP --dport 80 -j DNAT --to-destination SERVICE_IP:SERVICE_PORT
`
)

// CommandGenerator ...
type CommandGenerator struct {
}

// NewCommandGenerator ...
func NewCommandGenerator() *CommandGenerator {
	return &CommandGenerator{}
}

// GenerateCommandsForApplicationServers ...
func (commandGenerator *CommandGenerator) GenerateCommandsForApplicationServers(serviceInfo *domain.ServiceInfo,
	eventUUID string) error {
	for _, applicationServer := range serviceInfo.ApplicationServers {
		err := enrichApplicationServersByCommands(serviceInfo.ServiceIP,
			serviceInfo.ServicePort,
			serviceInfo.Healthcheck.Type,
			applicationServer,
			eventUUID)
		if err != nil {
			return fmt.Errorf("can't generate command for application server %v, got error: %v",
				err, applicationServer.ServerIP)
		}
	}
	return nil
}

func enrichApplicationServersByCommands(serviceIP,
	servicePort,
	serviceHealthcheckType string,
	applicationServer *domain.ApplicationServer,
	eventUUID string) error {
	switch serviceHealthcheckType {
	case "tcp":
		applicationServerIP := applicationServer.ServerIP
		applicationServerPort := (strings.Split(applicationServer.ServerHealthcheck.HealthcheckAddress, ":"))[1]
		tcpCommands := strings.ReplaceAll(rawCommandsForTCPHealthchecks,
			"SERVICE_IP", serviceIP)
		tcpCommands = strings.ReplaceAll(tcpCommands,
			"SERVICE_PORT", servicePort)
		tcpCommands = strings.ReplaceAll(tcpCommands,
			"APPLICATION_SERVER_IP", applicationServerIP)
		tcpCommands = strings.ReplaceAll(tcpCommands,
			"APPLICATION_SERVER_HEALTHCHECK_PORT", applicationServerPort)

		applicationServer.ServerСonfigurationCommands = tcpCommands
	case "icmp":
		applicationServerIP := applicationServer.ServerIP
		icmpCommands := strings.ReplaceAll(rawCommandsForICMPHealthchecks,
			"SERVICE_IP", serviceIP)
		icmpCommands = strings.ReplaceAll(icmpCommands,
			"SERVICE_PORT", servicePort)
		icmpCommands = strings.ReplaceAll(icmpCommands,
			"APPLICATION_SERVER_IP", applicationServerIP)

		applicationServer.ServerСonfigurationCommands = icmpCommands
	case "http", "httpAdvanced":
		applicationServerIP := applicationServer.ServerIP
		httpCommands := strings.ReplaceAll(rawCommandsForHTTPHealthchecks,
			"SERVICE_IP", serviceIP)
		httpCommands = strings.ReplaceAll(httpCommands,
			"SERVICE_PORT", servicePort)
		httpCommands = strings.ReplaceAll(httpCommands,
			"APPLICATION_SERVER_IP", applicationServerIP)

		applicationServer.ServerСonfigurationCommands = httpCommands
	default:
		return fmt.Errorf("unknown type for generate command: %v", serviceHealthcheckType)
	}
	return nil
}

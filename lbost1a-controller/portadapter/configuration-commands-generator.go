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
	eventID string) error {
	serviceInfo.RWMutex.RLock()
	defer serviceInfo.RWMutex.RUnlock()
	for _, applicationServer := range serviceInfo.ApplicationServers {
		if err := enrichApplicationServersByCommands(serviceInfo.IP,
			serviceInfo.Port,
			serviceInfo.HCType,
			applicationServer,
			eventID); err != nil {
			return fmt.Errorf("can't generate command for application server %v, got error: %v",
				err, applicationServer.IP)
		}
	}
	return nil
}

func enrichApplicationServersByCommands(serviceIP,
	servicePort,
	serviceHealthcheckType string,
	applicationServer *domain.ApplicationServer,
	eventID string) error {
	switch serviceHealthcheckType {
	case "tcp":
		applicationServerIP := applicationServer.IP
		healthcheckAddress := strings.Split(applicationServer.HCAddress, ":")
		applicationServerPort := ""
		if len(healthcheckAddress) > 1 {
			applicationServerPort = healthcheckAddress[1]
		}
		tcpCommands := strings.ReplaceAll(rawCommandsForTCPHealthchecks,
			"SERVICE_IP", serviceIP)
		tcpCommands = strings.ReplaceAll(tcpCommands,
			"SERVICE_PORT", servicePort)
		tcpCommands = strings.ReplaceAll(tcpCommands,
			"APPLICATION_SERVER_IP", applicationServerIP)
		tcpCommands = strings.ReplaceAll(tcpCommands,
			"APPLICATION_SERVER_HEALTHCHECK_PORT", applicationServerPort)

		applicationServer.ExampleBashCommands = tcpCommands
		return nil
	case "icmp":
		applicationServerIP := applicationServer.IP
		icmpCommands := strings.ReplaceAll(rawCommandsForICMPHealthchecks,
			"SERVICE_IP", serviceIP)
		icmpCommands = strings.ReplaceAll(icmpCommands,
			"SERVICE_PORT", servicePort)
		icmpCommands = strings.ReplaceAll(icmpCommands,
			"APPLICATION_SERVER_IP", applicationServerIP)

		applicationServer.ExampleBashCommands = icmpCommands
		return nil
	case "http", "httpAdvanced":
		applicationServerIP := applicationServer.IP
		httpCommands := strings.ReplaceAll(rawCommandsForHTTPHealthchecks,
			"SERVICE_IP", serviceIP)
		httpCommands = strings.ReplaceAll(httpCommands,
			"SERVICE_PORT", servicePort)
		httpCommands = strings.ReplaceAll(httpCommands,
			"APPLICATION_SERVER_IP", applicationServerIP)

		applicationServer.ExampleBashCommands = httpCommands
		return nil
	default:
		return fmt.Errorf("unknown type for generate command: %v", serviceHealthcheckType)
	}
}

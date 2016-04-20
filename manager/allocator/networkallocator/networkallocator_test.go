package networkallocator

import (
	"net"
	"testing"

	"github.com/docker/swarm-v2/api"
	"github.com/stretchr/testify/assert"
)

func newNetworkAllocator(t *testing.T) *NetworkAllocator {
	na, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, na)
	return na
}

func TestNew(t *testing.T) {
	newNetworkAllocator(t)
}

func TestAllocateInvalidIPAM(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
			DriverConfiguration: &api.Driver{},
			IPAM: &api.IPAMOptions{
				Driver: &api.Driver{
					Name: "invalidipam,",
				},
			},
		},
	}
	err := na.Allocate(n)
	assert.Error(t, err)
}

func TestAllocateInvalidDriver(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
			DriverConfiguration: &api.Driver{
				Name: "invaliddriver",
			},
		},
	}

	err := na.Allocate(n)
	assert.Error(t, err)
}

func TestNetworkDoubleAllocate(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
		},
	}

	err := na.Allocate(n)
	assert.NoError(t, err)

	err = na.Allocate(n)
	assert.Error(t, err)
}

func TestAllocateEmptyConfig(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
		},
	}

	err := na.Allocate(n)
	assert.NoError(t, err)
	assert.NotEqual(t, n.IPAM.Configurations, nil)
	assert.Equal(t, len(n.IPAM.Configurations), 1)
	assert.Equal(t, n.IPAM.Configurations[0].Range, "")
	assert.Equal(t, len(n.IPAM.Configurations[0].Reserved), 0)

	_, _, err = net.ParseCIDR(n.IPAM.Configurations[0].Subnet)
	assert.NoError(t, err)

	ip := net.ParseIP(n.IPAM.Configurations[0].Gateway)
	assert.NotEqual(t, ip, nil)
}

func TestAllocateWithOneSubnet(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
			DriverConfiguration: &api.Driver{},
			IPAM: &api.IPAMOptions{
				Driver: &api.Driver{},
				Configurations: []*api.IPAMConfiguration{
					{
						Subnet: "192.168.1.0/24",
					},
				},
			},
		},
	}

	err := na.Allocate(n)
	assert.NoError(t, err)
	assert.Equal(t, len(n.IPAM.Configurations), 1)
	assert.Equal(t, n.IPAM.Configurations[0].Range, "")
	assert.Equal(t, len(n.IPAM.Configurations[0].Reserved), 0)
	assert.Equal(t, n.IPAM.Configurations[0].Subnet, "192.168.1.0/24")

	ip := net.ParseIP(n.IPAM.Configurations[0].Gateway)
	assert.NotEqual(t, ip, nil)
}

func TestAllocateWithOneSubnetGateway(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
			DriverConfiguration: &api.Driver{},
			IPAM: &api.IPAMOptions{
				Driver: &api.Driver{},
				Configurations: []*api.IPAMConfiguration{
					{
						Subnet:  "192.168.1.0/24",
						Gateway: "192.168.1.1",
					},
				},
			},
		},
	}

	err := na.Allocate(n)
	assert.NoError(t, err)
	assert.Equal(t, len(n.IPAM.Configurations), 1)
	assert.Equal(t, n.IPAM.Configurations[0].Range, "")
	assert.Equal(t, len(n.IPAM.Configurations[0].Reserved), 0)
	assert.Equal(t, n.IPAM.Configurations[0].Subnet, "192.168.1.0/24")
	assert.Equal(t, n.IPAM.Configurations[0].Gateway, "192.168.1.1")
}

func TestAllocateWithOneSubnetInvalidGateway(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
			DriverConfiguration: &api.Driver{},
			IPAM: &api.IPAMOptions{
				Driver: &api.Driver{},
				Configurations: []*api.IPAMConfiguration{
					{
						Subnet:  "192.168.1.0/24",
						Gateway: "192.168.2.1",
					},
				},
			},
		},
	}

	err := na.Allocate(n)
	assert.Error(t, err)
}

func TestAllocateWithInvalidSubnet(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
			DriverConfiguration: &api.Driver{},
			IPAM: &api.IPAMOptions{
				Driver: &api.Driver{},
				Configurations: []*api.IPAMConfiguration{
					{
						Subnet: "1.1.1.1/32",
					},
				},
			},
		},
	}

	err := na.Allocate(n)
	assert.Error(t, err)
}

func TestAllocateWithTwoSubnetsNoGateway(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
			DriverConfiguration: &api.Driver{},
			IPAM: &api.IPAMOptions{
				Driver: &api.Driver{},
				Configurations: []*api.IPAMConfiguration{
					{
						Subnet: "192.168.1.0/24",
					},
					{
						Subnet: "192.168.2.0/24",
					},
				},
			},
		},
	}

	err := na.Allocate(n)
	assert.NoError(t, err)
	assert.Equal(t, len(n.IPAM.Configurations), 2)
	assert.Equal(t, n.IPAM.Configurations[0].Range, "")
	assert.Equal(t, len(n.IPAM.Configurations[0].Reserved), 0)
	assert.Equal(t, n.IPAM.Configurations[0].Subnet, "192.168.1.0/24")
	assert.Equal(t, n.IPAM.Configurations[1].Range, "")
	assert.Equal(t, len(n.IPAM.Configurations[1].Reserved), 0)
	assert.Equal(t, n.IPAM.Configurations[1].Subnet, "192.168.2.0/24")

	ip := net.ParseIP(n.IPAM.Configurations[0].Gateway)
	assert.NotEqual(t, ip, nil)
	ip = net.ParseIP(n.IPAM.Configurations[1].Gateway)
	assert.NotEqual(t, ip, nil)
}

func TestFree(t *testing.T) {
	na := newNetworkAllocator(t)
	n := &api.Network{
		ID: "testID",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test",
			},
			DriverConfiguration: &api.Driver{},
			IPAM: &api.IPAMOptions{
				Driver: &api.Driver{},
				Configurations: []*api.IPAMConfiguration{
					{
						Subnet:  "192.168.1.0/24",
						Gateway: "192.168.1.1",
					},
				},
			},
		},
	}

	err := na.Allocate(n)
	assert.NoError(t, err)

	err = na.Deallocate(n)
	assert.NoError(t, err)

	// Reallocate again to make sure it succeeds.
	err = na.Allocate(n)
	assert.NoError(t, err)
}

func TestAllocateTaskFree(t *testing.T) {
	na := newNetworkAllocator(t)
	n1 := &api.Network{
		ID: "testID1",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test1",
			},
			DriverConfiguration: &api.Driver{},
			IPAM: &api.IPAMOptions{
				Driver: &api.Driver{},
				Configurations: []*api.IPAMConfiguration{
					{
						Subnet:  "192.168.1.0/24",
						Gateway: "192.168.1.1",
					},
				},
			},
		},
	}

	n2 := &api.Network{
		ID: "testID2",
		Spec: &api.NetworkSpec{
			Annotations: api.Annotations{
				Name: "test2",
			},
			DriverConfiguration: &api.Driver{},
			IPAM: &api.IPAMOptions{
				Driver: &api.Driver{},
				Configurations: []*api.IPAMConfiguration{
					{
						Subnet:  "192.168.2.0/24",
						Gateway: "192.168.2.1",
					},
				},
			},
		},
	}

	task := &api.Task{
		Networks: []*api.Task_NetworkAttachment{
			{
				Network: n1,
			},
			{
				Network: n2,
			},
		},
	}

	err := na.Allocate(n1)
	assert.NoError(t, err)

	err = na.Allocate(n2)
	assert.NoError(t, err)

	err = na.AllocateTask(task)
	assert.NoError(t, err)
	assert.Equal(t, len(task.Networks[0].Addresses), 1)
	assert.Equal(t, len(task.Networks[1].Addresses), 1)

	_, subnet1, _ := net.ParseCIDR("192.168.1.0/24")
	_, subnet2, _ := net.ParseCIDR("192.168.2.0/24")

	ip1, _, err := net.ParseCIDR(task.Networks[0].Addresses[0])
	assert.NoError(t, err)

	ip2, _, err := net.ParseCIDR(task.Networks[1].Addresses[0])
	assert.NoError(t, err)

	assert.Equal(t, subnet1.Contains(ip1), true)
	assert.Equal(t, subnet2.Contains(ip2), true)

	err = na.DeallocateTask(task)
	assert.NoError(t, err)
	assert.Equal(t, len(task.Networks[0].Addresses), 0)
	assert.Equal(t, len(task.Networks[1].Addresses), 0)

	// Try allocation after free
	err = na.AllocateTask(task)
	assert.NoError(t, err)
	assert.Equal(t, len(task.Networks[0].Addresses), 1)
	assert.Equal(t, len(task.Networks[1].Addresses), 1)

	ip1, _, err = net.ParseCIDR(task.Networks[0].Addresses[0])
	assert.NoError(t, err)

	ip2, _, err = net.ParseCIDR(task.Networks[1].Addresses[0])
	assert.NoError(t, err)

	assert.Equal(t, subnet1.Contains(ip1), true)
	assert.Equal(t, subnet2.Contains(ip2), true)

	err = na.DeallocateTask(task)
	assert.NoError(t, err)
	assert.Equal(t, len(task.Networks[0].Addresses), 0)
	assert.Equal(t, len(task.Networks[1].Addresses), 0)

	// Try to free endpoints on an already freed task
	err = na.DeallocateTask(task)
	assert.NoError(t, err)
}

func TestServiceAllocate(t *testing.T) {
	na := newNetworkAllocator(t)
	s := &api.Service{
		ID: "testID1",
		Spec: &api.ServiceSpec{
			Endpoint: &api.Endpoint{
				Ports: []*api.Endpoint_PortConfiguration{
					{
						Name: "http",
						Port: 80,
					},
					{
						Name: "https",
						Port: 443,
					},
				},
			},
		},
	}

	err := na.ServiceAllocate(s)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(s.Endpoint.Ports))
	assert.True(t, s.Endpoint.Ports[0].NodePort >= dynamicPortStart &&
		s.Endpoint.Ports[0].NodePort <= dynamicPortEnd)
	assert.True(t, s.Endpoint.Ports[1].NodePort >= dynamicPortStart &&
		s.Endpoint.Ports[1].NodePort <= dynamicPortEnd)
}

func TestServiceAllocateUserDefinedPorts(t *testing.T) {
	na := newNetworkAllocator(t)
	s := &api.Service{
		ID: "testID1",
		Spec: &api.ServiceSpec{
			Endpoint: &api.Endpoint{
				Ports: []*api.Endpoint_PortConfiguration{
					{
						Name:     "some_tcp",
						Port:     1234,
						NodePort: 1234,
					},
					{
						Name:     "some_udp",
						Port:     1234,
						NodePort: 1234,
						Protocol: api.Endpoint_UDP,
					},
				},
			},
		},
	}

	err := na.ServiceAllocate(s)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(s.Endpoint.Ports))
	assert.Equal(t, uint32(1234), s.Endpoint.Ports[0].NodePort)
	assert.Equal(t, uint32(1234), s.Endpoint.Ports[1].NodePort)
}

func TestServiceAllocateConflictingUserDefinedPorts(t *testing.T) {
	na := newNetworkAllocator(t)
	s := &api.Service{
		ID: "testID1",
		Spec: &api.ServiceSpec{
			Endpoint: &api.Endpoint{
				Ports: []*api.Endpoint_PortConfiguration{
					{
						Name:     "some_tcp",
						Port:     1234,
						NodePort: 1234,
					},
					{
						Name:     "some_other_tcp",
						Port:     1234,
						NodePort: 1234,
					},
				},
			},
		},
	}

	err := na.ServiceAllocate(s)
	assert.Error(t, err)
}

func TestServiceDeallocateAllocate(t *testing.T) {
	na := newNetworkAllocator(t)
	s := &api.Service{
		ID: "testID1",
		Spec: &api.ServiceSpec{
			Endpoint: &api.Endpoint{
				Ports: []*api.Endpoint_PortConfiguration{
					{
						Name:     "some_tcp",
						Port:     1234,
						NodePort: 1234,
					},
				},
			},
		},
	}

	err := na.ServiceAllocate(s)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(s.Endpoint.Ports))
	assert.Equal(t, uint32(1234), s.Endpoint.Ports[0].NodePort)

	err = na.ServiceDeallocate(s)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(s.Endpoint.Ports))

	// Allocate again.
	err = na.ServiceAllocate(s)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(s.Endpoint.Ports))
	assert.Equal(t, uint32(1234), s.Endpoint.Ports[0].NodePort)
}

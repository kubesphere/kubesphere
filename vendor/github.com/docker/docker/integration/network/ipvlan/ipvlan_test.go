package ipvlan

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	dclient "github.com/docker/docker/client"
	"github.com/docker/docker/integration/internal/container"
	n "github.com/docker/docker/integration/network"
	"github.com/docker/docker/internal/test/daemon"
	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/skip"
)

func TestDockerNetworkIpvlanPersistance(t *testing.T) {
	// verify the driver automatically provisions the 802.1q link (di-dummy0.70)
	skip.If(t, testEnv.DaemonInfo.OSType != "linux")
	skip.If(t, testEnv.IsRemoteDaemon())
	skip.If(t, !ipvlanKernelSupport(), "Kernel doesn't support ipvlan")

	d := daemon.New(t, daemon.WithExperimental)
	d.StartWithBusybox(t)
	defer d.Stop(t)

	// master dummy interface 'di' notation represent 'docker ipvlan'
	master := "di-dummy0"
	n.CreateMasterDummy(t, master)
	defer n.DeleteInterface(t, master)

	client, err := d.NewClient()
	assert.NilError(t, err)

	// create a network specifying the desired sub-interface name
	_, err = client.NetworkCreate(context.Background(), "di-persist", types.NetworkCreate{
		Driver: "ipvlan",
		Options: map[string]string{
			"parent": "di-dummy0.70",
		},
	})
	assert.NilError(t, err)
	assert.Check(t, n.IsNetworkAvailable(client, "di-persist"))
	// Restart docker daemon to test the config has persisted to disk
	d.Restart(t)
	assert.Check(t, n.IsNetworkAvailable(client, "di-persist"))
}

func TestDockerNetworkIpvlan(t *testing.T) {
	skip.If(t, testEnv.DaemonInfo.OSType != "linux")
	skip.If(t, testEnv.IsRemoteDaemon())
	skip.If(t, !ipvlanKernelSupport(), "Kernel doesn't support ipvlan")

	for _, tc := range []struct {
		name string
		test func(dclient.APIClient) func(*testing.T)
	}{
		{
			name: "Subinterface",
			test: testIpvlanSubinterface,
		}, {
			name: "OverlapParent",
			test: testIpvlanOverlapParent,
		}, {
			name: "L2NilParent",
			test: testIpvlanL2NilParent,
		}, {
			name: "L2InternalMode",
			test: testIpvlanL2InternalMode,
		}, {
			name: "L3NilParent",
			test: testIpvlanL3NilParent,
		}, {
			name: "L3InternalMode",
			test: testIpvlanL3InternalMode,
		}, {
			name: "L2MultiSubnet",
			test: testIpvlanL2MultiSubnet,
		}, {
			name: "L3MultiSubnet",
			test: testIpvlanL3MultiSubnet,
		}, {
			name: "Addressing",
			test: testIpvlanAddressing,
		},
	} {
		d := daemon.New(t, daemon.WithExperimental)
		d.StartWithBusybox(t)

		client, err := d.NewClient()
		assert.NilError(t, err)

		t.Run(tc.name, tc.test(client))

		d.Stop(t)
		// FIXME(vdemeester) clean network
	}
}

func testIpvlanSubinterface(client dclient.APIClient) func(*testing.T) {
	return func(t *testing.T) {
		master := "di-dummy0"
		n.CreateMasterDummy(t, master)
		defer n.DeleteInterface(t, master)

		_, err := client.NetworkCreate(context.Background(), "di-subinterface", types.NetworkCreate{
			Driver: "ipvlan",
			Options: map[string]string{
				"parent": "di-dummy0.60",
			},
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "di-subinterface"))

		// delete the network while preserving the parent link
		err = client.NetworkRemove(context.Background(), "di-subinterface")
		assert.NilError(t, err)

		assert.Check(t, n.IsNetworkNotAvailable(client, "di-subinterface"))
		// verify the network delete did not delete the predefined link
		n.LinkExists(t, "di-dummy0")
	}
}

func testIpvlanOverlapParent(client dclient.APIClient) func(*testing.T) {
	return func(t *testing.T) {
		// verify the same parent interface cannot be used if already in use by an existing network
		master := "di-dummy0"
		n.CreateMasterDummy(t, master)
		defer n.DeleteInterface(t, master)
		n.CreateVlanInterface(t, master, "di-dummy0.30", "30")

		_, err := client.NetworkCreate(context.Background(), "di-subinterface", types.NetworkCreate{
			Driver: "ipvlan",
			Options: map[string]string{
				"parent": "di-dummy0.30",
			},
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "di-subinterface"))

		_, err = client.NetworkCreate(context.Background(), "di-subinterface", types.NetworkCreate{
			Driver: "ipvlan",
			Options: map[string]string{
				"parent": "di-dummy0.30",
			},
		})
		// verify that the overlap returns an error
		assert.Check(t, err != nil)
	}
}

func testIpvlanL2NilParent(client dclient.APIClient) func(*testing.T) {
	return func(t *testing.T) {
		// ipvlan l2 mode - dummy parent interface is provisioned dynamically
		_, err := client.NetworkCreate(context.Background(), "di-nil-parent", types.NetworkCreate{
			Driver: "ipvlan",
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "di-nil-parent"))

		ctx := context.Background()
		id1 := container.Run(t, ctx, client, container.WithNetworkMode("di-nil-parent"))
		id2 := container.Run(t, ctx, client, container.WithNetworkMode("di-nil-parent"))

		_, err = container.Exec(ctx, client, id2, []string{"ping", "-c", "1", id1})
		assert.NilError(t, err)
	}
}

func testIpvlanL2InternalMode(client dclient.APIClient) func(*testing.T) {
	return func(t *testing.T) {
		_, err := client.NetworkCreate(context.Background(), "di-internal", types.NetworkCreate{
			Driver:   "ipvlan",
			Internal: true,
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "di-internal"))

		ctx := context.Background()
		id1 := container.Run(t, ctx, client, container.WithNetworkMode("di-internal"))
		id2 := container.Run(t, ctx, client, container.WithNetworkMode("di-internal"))

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, err = container.Exec(timeoutCtx, client, id1, []string{"ping", "-c", "1", "-w", "1", "8.8.8.8"})
		// FIXME(vdemeester) check the time of error ?
		assert.Check(t, err != nil)
		assert.Check(t, timeoutCtx.Err() == context.DeadlineExceeded)

		_, err = container.Exec(ctx, client, id2, []string{"ping", "-c", "1", id1})
		assert.NilError(t, err)
	}
}

func testIpvlanL3NilParent(client dclient.APIClient) func(*testing.T) {
	return func(t *testing.T) {
		_, err := client.NetworkCreate(context.Background(), "di-nil-parent-l3", types.NetworkCreate{
			Driver: "ipvlan",
			Options: map[string]string{
				"ipvlan_mode": "l3",
			},
			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{
					{
						Subnet:     "172.28.230.0/24",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "172.28.220.0/24",
						AuxAddress: map[string]string{},
					},
				},
			},
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "di-nil-parent-l3"))

		ctx := context.Background()
		id1 := container.Run(t, ctx, client,
			container.WithNetworkMode("di-nil-parent-l3"),
			container.WithIPv4("di-nil-parent-l3", "172.28.220.10"),
		)
		id2 := container.Run(t, ctx, client,
			container.WithNetworkMode("di-nil-parent-l3"),
			container.WithIPv4("di-nil-parent-l3", "172.28.230.10"),
		)

		_, err = container.Exec(ctx, client, id2, []string{"ping", "-c", "1", id1})
		assert.NilError(t, err)
	}
}

func testIpvlanL3InternalMode(client dclient.APIClient) func(*testing.T) {
	return func(t *testing.T) {
		_, err := client.NetworkCreate(context.Background(), "di-internal-l3", types.NetworkCreate{
			Driver:   "ipvlan",
			Internal: true,
			Options: map[string]string{
				"ipvlan_mode": "l3",
			},
			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{
					{
						Subnet:     "172.28.230.0/24",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "172.28.220.0/24",
						AuxAddress: map[string]string{},
					},
				},
			},
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "di-internal-l3"))

		ctx := context.Background()
		id1 := container.Run(t, ctx, client,
			container.WithNetworkMode("di-internal-l3"),
			container.WithIPv4("di-internal-l3", "172.28.220.10"),
		)
		id2 := container.Run(t, ctx, client,
			container.WithNetworkMode("di-internal-l3"),
			container.WithIPv4("di-internal-l3", "172.28.230.10"),
		)

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, err = container.Exec(timeoutCtx, client, id1, []string{"ping", "-c", "1", "-w", "1", "8.8.8.8"})
		// FIXME(vdemeester) check the time of error ?
		assert.Check(t, err != nil)
		assert.Check(t, timeoutCtx.Err() == context.DeadlineExceeded)

		_, err = container.Exec(ctx, client, id2, []string{"ping", "-c", "1", id1})
		assert.NilError(t, err)
	}
}

func testIpvlanL2MultiSubnet(client dclient.APIClient) func(*testing.T) {
	return func(t *testing.T) {
		_, err := client.NetworkCreate(context.Background(), "dualstackl2", types.NetworkCreate{
			Driver:     "ipvlan",
			EnableIPv6: true,
			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{
					{
						Subnet:     "172.28.200.0/24",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "172.28.202.0/24",
						Gateway:    "172.28.202.254",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "2001:db8:abc8::/64",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "2001:db8:abc6::/64",
						Gateway:    "2001:db8:abc6::254",
						AuxAddress: map[string]string{},
					},
				},
			},
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "dualstackl2"))

		// start dual stack containers and verify the user specified --ip and --ip6 addresses on subnets 172.28.100.0/24 and 2001:db8:abc2::/64
		ctx := context.Background()
		id1 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl2"),
			container.WithIPv4("dualstackl2", "172.28.200.20"),
			container.WithIPv6("dualstackl2", "2001:db8:abc8::20"),
		)
		id2 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl2"),
			container.WithIPv4("dualstackl2", "172.28.200.21"),
			container.WithIPv6("dualstackl2", "2001:db8:abc8::21"),
		)
		c1, err := client.ContainerInspect(ctx, id1)
		assert.NilError(t, err)

		// verify ipv4 connectivity to the explicit --ipv address second to first
		_, err = container.Exec(ctx, client, id2, []string{"ping", "-c", "1", c1.NetworkSettings.Networks["dualstackl2"].IPAddress})
		assert.NilError(t, err)
		// verify ipv6 connectivity to the explicit --ipv6 address second to first
		_, err = container.Exec(ctx, client, id2, []string{"ping6", "-c", "1", c1.NetworkSettings.Networks["dualstackl2"].GlobalIPv6Address})
		assert.NilError(t, err)

		// start dual stack containers and verify the user specified --ip and --ip6 addresses on subnets 172.28.102.0/24 and 2001:db8:abc4::/64
		id3 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl2"),
			container.WithIPv4("dualstackl2", "172.28.202.20"),
			container.WithIPv6("dualstackl2", "2001:db8:abc6::20"),
		)
		id4 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl2"),
			container.WithIPv4("dualstackl2", "172.28.202.21"),
			container.WithIPv6("dualstackl2", "2001:db8:abc6::21"),
		)
		c3, err := client.ContainerInspect(ctx, id3)
		assert.NilError(t, err)

		// verify ipv4 connectivity to the explicit --ipv address from third to fourth
		_, err = container.Exec(ctx, client, id4, []string{"ping", "-c", "1", c3.NetworkSettings.Networks["dualstackl2"].IPAddress})
		assert.NilError(t, err)
		// verify ipv6 connectivity to the explicit --ipv6 address from third to fourth
		_, err = container.Exec(ctx, client, id4, []string{"ping6", "-c", "1", c3.NetworkSettings.Networks["dualstackl2"].GlobalIPv6Address})
		assert.NilError(t, err)

		// Inspect the v4 gateway to ensure the proper default GW was assigned
		assert.Equal(t, c1.NetworkSettings.Networks["dualstackl2"].Gateway, "172.28.200.1")
		// Inspect the v6 gateway to ensure the proper default GW was assigned
		assert.Equal(t, c1.NetworkSettings.Networks["dualstackl2"].IPv6Gateway, "2001:db8:abc8::1")
		// Inspect the v4 gateway to ensure the proper explicitly assigned default GW was assigned
		assert.Equal(t, c3.NetworkSettings.Networks["dualstackl2"].Gateway, "172.28.202.254")
		// Inspect the v6 gateway to ensure the proper explicitly assigned default GW was assigned
		assert.Equal(t, c3.NetworkSettings.Networks["dualstackl2"].IPv6Gateway, "2001:db8:abc6::254")
	}
}

func testIpvlanL3MultiSubnet(client dclient.APIClient) func(*testing.T) {
	return func(t *testing.T) {
		_, err := client.NetworkCreate(context.Background(), "dualstackl3", types.NetworkCreate{
			Driver:     "ipvlan",
			EnableIPv6: true,
			Options: map[string]string{
				"ipvlan_mode": "l3",
			},
			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{
					{
						Subnet:     "172.28.10.0/24",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "172.28.12.0/24",
						Gateway:    "172.28.12.254",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "2001:db8:abc9::/64",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "2001:db8:abc7::/64",
						Gateway:    "2001:db8:abc7::254",
						AuxAddress: map[string]string{},
					},
				},
			},
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "dualstackl3"))

		// start dual stack containers and verify the user specified --ip and --ip6 addresses on subnets 172.28.100.0/24 and 2001:db8:abc2::/64
		ctx := context.Background()
		id1 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl3"),
			container.WithIPv4("dualstackl3", "172.28.10.20"),
			container.WithIPv6("dualstackl3", "2001:db8:abc9::20"),
		)
		id2 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl3"),
			container.WithIPv4("dualstackl3", "172.28.10.21"),
			container.WithIPv6("dualstackl3", "2001:db8:abc9::21"),
		)
		c1, err := client.ContainerInspect(ctx, id1)
		assert.NilError(t, err)

		// verify ipv4 connectivity to the explicit --ipv address second to first
		_, err = container.Exec(ctx, client, id2, []string{"ping", "-c", "1", c1.NetworkSettings.Networks["dualstackl3"].IPAddress})
		assert.NilError(t, err)
		// verify ipv6 connectivity to the explicit --ipv6 address second to first
		_, err = container.Exec(ctx, client, id2, []string{"ping6", "-c", "1", c1.NetworkSettings.Networks["dualstackl3"].GlobalIPv6Address})
		assert.NilError(t, err)

		// start dual stack containers and verify the user specified --ip and --ip6 addresses on subnets 172.28.102.0/24 and 2001:db8:abc4::/64
		id3 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl3"),
			container.WithIPv4("dualstackl3", "172.28.12.20"),
			container.WithIPv6("dualstackl3", "2001:db8:abc7::20"),
		)
		id4 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl3"),
			container.WithIPv4("dualstackl3", "172.28.12.21"),
			container.WithIPv6("dualstackl3", "2001:db8:abc7::21"),
		)
		c3, err := client.ContainerInspect(ctx, id3)
		assert.NilError(t, err)

		// verify ipv4 connectivity to the explicit --ipv address from third to fourth
		_, err = container.Exec(ctx, client, id4, []string{"ping", "-c", "1", c3.NetworkSettings.Networks["dualstackl3"].IPAddress})
		assert.NilError(t, err)
		// verify ipv6 connectivity to the explicit --ipv6 address from third to fourth
		_, err = container.Exec(ctx, client, id4, []string{"ping6", "-c", "1", c3.NetworkSettings.Networks["dualstackl3"].GlobalIPv6Address})
		assert.NilError(t, err)

		// Inspect the v4 gateway to ensure no next hop is assigned in L3 mode
		assert.Equal(t, c1.NetworkSettings.Networks["dualstackl3"].Gateway, "")
		// Inspect the v6 gateway to ensure the explicitly specified default GW is ignored per L3 mode enabled
		assert.Equal(t, c1.NetworkSettings.Networks["dualstackl3"].IPv6Gateway, "")
		// Inspect the v4 gateway to ensure no next hop is assigned in L3 mode
		assert.Equal(t, c3.NetworkSettings.Networks["dualstackl3"].Gateway, "")
		// Inspect the v6 gateway to ensure the explicitly specified default GW is ignored per L3 mode enabled
		assert.Equal(t, c3.NetworkSettings.Networks["dualstackl3"].IPv6Gateway, "")
	}
}

func testIpvlanAddressing(client dclient.APIClient) func(*testing.T) {
	return func(t *testing.T) {
		// Verify ipvlan l2 mode sets the proper default gateway routes via netlink
		// for either an explicitly set route by the user or inferred via default IPAM
		_, err := client.NetworkCreate(context.Background(), "dualstackl2", types.NetworkCreate{
			Driver:     "ipvlan",
			EnableIPv6: true,
			Options: map[string]string{
				"ipvlan_mode": "l2",
			},
			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{
					{
						Subnet:     "172.28.140.0/24",
						Gateway:    "172.28.140.254",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "2001:db8:abcb::/64",
						AuxAddress: map[string]string{},
					},
				},
			},
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "dualstackl2"))

		ctx := context.Background()
		id1 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl2"),
		)
		// Validate ipvlan l2 mode defaults gateway sets the default IPAM next-hop inferred from the subnet
		result, err := container.Exec(ctx, client, id1, []string{"ip", "route"})
		assert.NilError(t, err)
		assert.Check(t, strings.Contains(result.Combined(), "default via 172.28.140.254 dev eth0"))
		// Validate ipvlan l2 mode sets the v6 gateway to the user specified default gateway/next-hop
		result, err = container.Exec(ctx, client, id1, []string{"ip", "-6", "route"})
		assert.NilError(t, err)
		assert.Check(t, strings.Contains(result.Combined(), "default via 2001:db8:abcb::1 dev eth0"))

		// Validate ipvlan l3 mode sets the v4 gateway to dev eth0 and disregards any explicit or inferred next-hops
		_, err = client.NetworkCreate(context.Background(), "dualstackl3", types.NetworkCreate{
			Driver:     "ipvlan",
			EnableIPv6: true,
			Options: map[string]string{
				"ipvlan_mode": "l3",
			},
			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{
					{
						Subnet:     "172.28.160.0/24",
						Gateway:    "172.28.160.254",
						AuxAddress: map[string]string{},
					},
					{
						Subnet:     "2001:db8:abcd::/64",
						Gateway:    "2001:db8:abcd::254",
						AuxAddress: map[string]string{},
					},
				},
			},
		})
		assert.NilError(t, err)
		assert.Check(t, n.IsNetworkAvailable(client, "dualstackl3"))

		id2 := container.Run(t, ctx, client,
			container.WithNetworkMode("dualstackl3"),
		)
		// Validate ipvlan l3 mode sets the v4 gateway to dev eth0 and disregards any explicit or inferred next-hops
		result, err = container.Exec(ctx, client, id2, []string{"ip", "route"})
		assert.NilError(t, err)
		assert.Check(t, strings.Contains(result.Combined(), "default dev eth0"))
		// Validate ipvlan l3 mode sets the v6 gateway to dev eth0 and disregards any explicit or inferred next-hops
		result, err = container.Exec(ctx, client, id2, []string{"ip", "-6", "route"})
		assert.NilError(t, err)
		assert.Check(t, strings.Contains(result.Combined(), "default dev eth0"))
	}
}

// ensure Kernel version is >= v4.2 for ipvlan support
func ipvlanKernelSupport() bool {
	return n.CheckKernelMajorVersionGreaterOrEqualThen(4, 2)
}

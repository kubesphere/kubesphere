// +build !windows

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/initca"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/integration-cli/checker"
	"github.com/docker/docker/integration-cli/daemon"
	testdaemon "github.com/docker/docker/internal/test/daemon"
	"github.com/docker/docker/internal/test/request"
	"github.com/docker/swarmkit/ca"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
)

var defaultReconciliationTimeout = 30 * time.Second

func (s *DockerSwarmSuite) TestAPISwarmInit(c *check.C) {
	// todo: should find a better way to verify that components are running than /info
	d1 := s.AddDaemon(c, true, true)
	info := d1.SwarmInfo(c)
	c.Assert(info.ControlAvailable, checker.True)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)
	c.Assert(info.Cluster.RootRotationInProgress, checker.False)

	d2 := s.AddDaemon(c, true, false)
	info = d2.SwarmInfo(c)
	c.Assert(info.ControlAvailable, checker.False)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)

	// Leaving cluster
	c.Assert(d2.SwarmLeave(false), checker.IsNil)

	info = d2.SwarmInfo(c)
	c.Assert(info.ControlAvailable, checker.False)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateInactive)

	d2.SwarmJoin(c, swarm.JoinRequest{
		ListenAddr:  d1.SwarmListenAddr(),
		JoinToken:   d1.JoinTokens(c).Worker,
		RemoteAddrs: []string{d1.SwarmListenAddr()},
	})

	info = d2.SwarmInfo(c)
	c.Assert(info.ControlAvailable, checker.False)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)

	// Current state restoring after restarts
	d1.Stop(c)
	d2.Stop(c)

	d1.Start(c)
	d2.Start(c)

	info = d1.SwarmInfo(c)
	c.Assert(info.ControlAvailable, checker.True)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)

	info = d2.SwarmInfo(c)
	c.Assert(info.ControlAvailable, checker.False)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)
}

func (s *DockerSwarmSuite) TestAPISwarmJoinToken(c *check.C) {
	d1 := s.AddDaemon(c, false, false)
	d1.SwarmInit(c, swarm.InitRequest{})

	// todo: error message differs depending if some components of token are valid

	d2 := s.AddDaemon(c, false, false)
	c2 := d2.NewClientT(c)
	err := c2.SwarmJoin(context.Background(), swarm.JoinRequest{
		ListenAddr:  d2.SwarmListenAddr(),
		RemoteAddrs: []string{d1.SwarmListenAddr()},
	})
	c.Assert(err, checker.NotNil)
	c.Assert(err.Error(), checker.Contains, "join token is necessary")
	info := d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateInactive)

	err = c2.SwarmJoin(context.Background(), swarm.JoinRequest{
		ListenAddr:  d2.SwarmListenAddr(),
		JoinToken:   "foobaz",
		RemoteAddrs: []string{d1.SwarmListenAddr()},
	})
	c.Assert(err, checker.NotNil)
	c.Assert(err.Error(), checker.Contains, "invalid join token")
	info = d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateInactive)

	workerToken := d1.JoinTokens(c).Worker

	d2.SwarmJoin(c, swarm.JoinRequest{
		ListenAddr:  d2.SwarmListenAddr(),
		JoinToken:   workerToken,
		RemoteAddrs: []string{d1.SwarmListenAddr()},
	})
	info = d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)
	c.Assert(d2.SwarmLeave(false), checker.IsNil)
	info = d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateInactive)

	// change tokens
	d1.RotateTokens(c)

	err = c2.SwarmJoin(context.Background(), swarm.JoinRequest{
		ListenAddr:  d2.SwarmListenAddr(),
		JoinToken:   workerToken,
		RemoteAddrs: []string{d1.SwarmListenAddr()},
	})
	c.Assert(err, checker.NotNil)
	c.Assert(err.Error(), checker.Contains, "join token is necessary")
	info = d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateInactive)

	workerToken = d1.JoinTokens(c).Worker

	d2.SwarmJoin(c, swarm.JoinRequest{JoinToken: workerToken, RemoteAddrs: []string{d1.SwarmListenAddr()}})
	info = d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)
	c.Assert(d2.SwarmLeave(false), checker.IsNil)
	info = d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateInactive)

	// change spec, don't change tokens
	d1.UpdateSwarm(c, func(s *swarm.Spec) {})

	err = c2.SwarmJoin(context.Background(), swarm.JoinRequest{
		ListenAddr:  d2.SwarmListenAddr(),
		RemoteAddrs: []string{d1.SwarmListenAddr()},
	})
	c.Assert(err, checker.NotNil)
	c.Assert(err.Error(), checker.Contains, "join token is necessary")
	info = d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateInactive)

	d2.SwarmJoin(c, swarm.JoinRequest{JoinToken: workerToken, RemoteAddrs: []string{d1.SwarmListenAddr()}})
	info = d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)
	c.Assert(d2.SwarmLeave(false), checker.IsNil)
	info = d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateInactive)
}

func (s *DockerSwarmSuite) TestUpdateSwarmAddExternalCA(c *check.C) {
	d1 := s.AddDaemon(c, false, false)
	d1.SwarmInit(c, swarm.InitRequest{})
	d1.UpdateSwarm(c, func(s *swarm.Spec) {
		s.CAConfig.ExternalCAs = []*swarm.ExternalCA{
			{
				Protocol: swarm.ExternalCAProtocolCFSSL,
				URL:      "https://thishasnoca.org",
			},
			{
				Protocol: swarm.ExternalCAProtocolCFSSL,
				URL:      "https://thishasacacert.org",
				CACert:   "cacert",
			},
		}
	})
	info := d1.SwarmInfo(c)
	c.Assert(info.Cluster.Spec.CAConfig.ExternalCAs, checker.HasLen, 2)
	c.Assert(info.Cluster.Spec.CAConfig.ExternalCAs[0].CACert, checker.Equals, "")
	c.Assert(info.Cluster.Spec.CAConfig.ExternalCAs[1].CACert, checker.Equals, "cacert")
}

func (s *DockerSwarmSuite) TestAPISwarmCAHash(c *check.C) {
	d1 := s.AddDaemon(c, true, true)
	d2 := s.AddDaemon(c, false, false)
	splitToken := strings.Split(d1.JoinTokens(c).Worker, "-")
	splitToken[2] = "1kxftv4ofnc6mt30lmgipg6ngf9luhwqopfk1tz6bdmnkubg0e"
	replacementToken := strings.Join(splitToken, "-")
	c2 := d2.NewClientT(c)
	err := c2.SwarmJoin(context.Background(), swarm.JoinRequest{
		ListenAddr:  d2.SwarmListenAddr(),
		JoinToken:   replacementToken,
		RemoteAddrs: []string{d1.SwarmListenAddr()},
	})
	c.Assert(err, checker.NotNil)
	c.Assert(err.Error(), checker.Contains, "remote CA does not match fingerprint")
}

func (s *DockerSwarmSuite) TestAPISwarmPromoteDemote(c *check.C) {
	d1 := s.AddDaemon(c, false, false)
	d1.SwarmInit(c, swarm.InitRequest{})
	d2 := s.AddDaemon(c, true, false)

	info := d2.SwarmInfo(c)
	c.Assert(info.ControlAvailable, checker.False)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)

	d1.UpdateNode(c, d2.NodeID(), func(n *swarm.Node) {
		n.Spec.Role = swarm.NodeRoleManager
	})

	waitAndAssert(c, defaultReconciliationTimeout, d2.CheckControlAvailable, checker.True)

	d1.UpdateNode(c, d2.NodeID(), func(n *swarm.Node) {
		n.Spec.Role = swarm.NodeRoleWorker
	})

	waitAndAssert(c, defaultReconciliationTimeout, d2.CheckControlAvailable, checker.False)

	// Wait for the role to change to worker in the cert. This is partially
	// done because it's something worth testing in its own right, and
	// partially because changing the role from manager to worker and then
	// back to manager quickly might cause the node to pause for awhile
	// while waiting for the role to change to worker, and the test can
	// time out during this interval.
	waitAndAssert(c, defaultReconciliationTimeout, func(c *check.C) (interface{}, check.CommentInterface) {
		certBytes, err := ioutil.ReadFile(filepath.Join(d2.Folder, "root", "swarm", "certificates", "swarm-node.crt"))
		if err != nil {
			return "", check.Commentf("error: %v", err)
		}
		certs, err := helpers.ParseCertificatesPEM(certBytes)
		if err == nil && len(certs) > 0 && len(certs[0].Subject.OrganizationalUnit) > 0 {
			return certs[0].Subject.OrganizationalUnit[0], nil
		}
		return "", check.Commentf("could not get organizational unit from certificate")
	}, checker.Equals, "swarm-worker")

	// Demoting last node should fail
	node := d1.GetNode(c, d1.NodeID())
	node.Spec.Role = swarm.NodeRoleWorker
	url := fmt.Sprintf("/nodes/%s/update?version=%d", node.ID, node.Version.Index)
	res, body, err := request.Post(url, request.Host(d1.Sock()), request.JSONBody(node.Spec))
	c.Assert(err, checker.IsNil)
	b, err := request.ReadBody(body)
	c.Assert(err, checker.IsNil)
	c.Assert(res.StatusCode, checker.Equals, http.StatusBadRequest, check.Commentf("output: %q", string(b)))

	// The warning specific to demoting the last manager is best-effort and
	// won't appear until the Role field of the demoted manager has been
	// updated.
	// Yes, I know this looks silly, but checker.Matches is broken, since
	// it anchors the regexp contrary to the documentation, and this makes
	// it impossible to match something that includes a line break.
	if !strings.Contains(string(b), "last manager of the swarm") {
		c.Assert(string(b), checker.Contains, "this would result in a loss of quorum")
	}
	info = d1.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)
	c.Assert(info.ControlAvailable, checker.True)

	// Promote already demoted node
	d1.UpdateNode(c, d2.NodeID(), func(n *swarm.Node) {
		n.Spec.Role = swarm.NodeRoleManager
	})

	waitAndAssert(c, defaultReconciliationTimeout, d2.CheckControlAvailable, checker.True)
}

func (s *DockerSwarmSuite) TestAPISwarmLeaderProxy(c *check.C) {
	// add three managers, one of these is leader
	d1 := s.AddDaemon(c, true, true)
	d2 := s.AddDaemon(c, true, true)
	d3 := s.AddDaemon(c, true, true)

	// start a service by hitting each of the 3 managers
	d1.CreateService(c, simpleTestService, func(s *swarm.Service) {
		s.Spec.Name = "test1"
	})
	d2.CreateService(c, simpleTestService, func(s *swarm.Service) {
		s.Spec.Name = "test2"
	})
	d3.CreateService(c, simpleTestService, func(s *swarm.Service) {
		s.Spec.Name = "test3"
	})

	// 3 services should be started now, because the requests were proxied to leader
	// query each node and make sure it returns 3 services
	for _, d := range []*daemon.Daemon{d1, d2, d3} {
		services := d.ListServices(c)
		c.Assert(services, checker.HasLen, 3)
	}
}

func (s *DockerSwarmSuite) TestAPISwarmLeaderElection(c *check.C) {
	// Create 3 nodes
	d1 := s.AddDaemon(c, true, true)
	d2 := s.AddDaemon(c, true, true)
	d3 := s.AddDaemon(c, true, true)

	// assert that the first node we made is the leader, and the other two are followers
	c.Assert(d1.GetNode(c, d1.NodeID()).ManagerStatus.Leader, checker.True)
	c.Assert(d1.GetNode(c, d2.NodeID()).ManagerStatus.Leader, checker.False)
	c.Assert(d1.GetNode(c, d3.NodeID()).ManagerStatus.Leader, checker.False)

	d1.Stop(c)

	var (
		leader    *daemon.Daemon   // keep track of leader
		followers []*daemon.Daemon // keep track of followers
	)
	checkLeader := func(nodes ...*daemon.Daemon) checkF {
		return func(c *check.C) (interface{}, check.CommentInterface) {
			// clear these out before each run
			leader = nil
			followers = nil
			for _, d := range nodes {
				if d.GetNode(c, d.NodeID()).ManagerStatus.Leader {
					leader = d
				} else {
					followers = append(followers, d)
				}
			}

			if leader == nil {
				return false, check.Commentf("no leader elected")
			}

			return true, check.Commentf("elected %v", leader.ID())
		}
	}

	// wait for an election to occur
	waitAndAssert(c, defaultReconciliationTimeout, checkLeader(d2, d3), checker.True)

	// assert that we have a new leader
	c.Assert(leader, checker.NotNil)

	// Keep track of the current leader, since we want that to be chosen.
	stableleader := leader

	// add the d1, the initial leader, back
	d1.Start(c)

	// TODO(stevvooe): may need to wait for rejoin here

	// wait for possible election
	waitAndAssert(c, defaultReconciliationTimeout, checkLeader(d1, d2, d3), checker.True)
	// pick out the leader and the followers again

	// verify that we still only have 1 leader and 2 followers
	c.Assert(leader, checker.NotNil)
	c.Assert(followers, checker.HasLen, 2)
	// and that after we added d1 back, the leader hasn't changed
	c.Assert(leader.NodeID(), checker.Equals, stableleader.NodeID())
}

func (s *DockerSwarmSuite) TestAPISwarmRaftQuorum(c *check.C) {
	d1 := s.AddDaemon(c, true, true)
	d2 := s.AddDaemon(c, true, true)
	d3 := s.AddDaemon(c, true, true)

	d1.CreateService(c, simpleTestService)

	d2.Stop(c)

	// make sure there is a leader
	waitAndAssert(c, defaultReconciliationTimeout, d1.CheckLeader, checker.IsNil)

	d1.CreateService(c, simpleTestService, func(s *swarm.Service) {
		s.Spec.Name = "top1"
	})

	d3.Stop(c)

	var service swarm.Service
	simpleTestService(&service)
	service.Spec.Name = "top2"
	cli, err := d1.NewClient()
	c.Assert(err, checker.IsNil)
	defer cli.Close()

	// d1 will eventually step down from leader because there is no longer an active quorum, wait for that to happen
	waitAndAssert(c, defaultReconciliationTimeout, func(c *check.C) (interface{}, check.CommentInterface) {
		_, err = cli.ServiceCreate(context.Background(), service.Spec, types.ServiceCreateOptions{})
		return err.Error(), nil
	}, checker.Contains, "Make sure more than half of the managers are online.")

	d2.Start(c)

	// make sure there is a leader
	waitAndAssert(c, defaultReconciliationTimeout, d1.CheckLeader, checker.IsNil)

	d1.CreateService(c, simpleTestService, func(s *swarm.Service) {
		s.Spec.Name = "top3"
	})
}

func (s *DockerSwarmSuite) TestAPISwarmLeaveRemovesContainer(c *check.C) {
	d := s.AddDaemon(c, true, true)

	instances := 2
	d.CreateService(c, simpleTestService, setInstances(instances))

	id, err := d.Cmd("run", "-d", "busybox", "top")
	c.Assert(err, checker.IsNil)
	id = strings.TrimSpace(id)

	waitAndAssert(c, defaultReconciliationTimeout, d.CheckActiveContainerCount, checker.Equals, instances+1)

	c.Assert(d.SwarmLeave(false), checker.NotNil)
	c.Assert(d.SwarmLeave(true), checker.IsNil)

	waitAndAssert(c, defaultReconciliationTimeout, d.CheckActiveContainerCount, checker.Equals, 1)

	id2, err := d.Cmd("ps", "-q")
	c.Assert(err, checker.IsNil)
	c.Assert(id, checker.HasPrefix, strings.TrimSpace(id2))
}

// #23629
func (s *DockerSwarmSuite) TestAPISwarmLeaveOnPendingJoin(c *check.C) {
	testRequires(c, Network)
	s.AddDaemon(c, true, true)
	d2 := s.AddDaemon(c, false, false)

	id, err := d2.Cmd("run", "-d", "busybox", "top")
	c.Assert(err, checker.IsNil)
	id = strings.TrimSpace(id)

	c2 := d2.NewClientT(c)
	err = c2.SwarmJoin(context.Background(), swarm.JoinRequest{
		ListenAddr:  d2.SwarmListenAddr(),
		RemoteAddrs: []string{"123.123.123.123:1234"},
	})
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), checker.Contains, "Timeout was reached")

	info := d2.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStatePending)

	c.Assert(d2.SwarmLeave(true), checker.IsNil)

	waitAndAssert(c, defaultReconciliationTimeout, d2.CheckActiveContainerCount, checker.Equals, 1)

	id2, err := d2.Cmd("ps", "-q")
	c.Assert(err, checker.IsNil)
	c.Assert(id, checker.HasPrefix, strings.TrimSpace(id2))
}

// #23705
func (s *DockerSwarmSuite) TestAPISwarmRestoreOnPendingJoin(c *check.C) {
	testRequires(c, Network)
	d := s.AddDaemon(c, false, false)
	client := d.NewClientT(c)
	err := client.SwarmJoin(context.Background(), swarm.JoinRequest{
		ListenAddr:  d.SwarmListenAddr(),
		RemoteAddrs: []string{"123.123.123.123:1234"},
	})
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), checker.Contains, "Timeout was reached")

	waitAndAssert(c, defaultReconciliationTimeout, d.CheckLocalNodeState, checker.Equals, swarm.LocalNodeStatePending)

	d.Stop(c)
	d.Start(c)

	info := d.SwarmInfo(c)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateInactive)
}

func (s *DockerSwarmSuite) TestAPISwarmManagerRestore(c *check.C) {
	d1 := s.AddDaemon(c, true, true)

	instances := 2
	id := d1.CreateService(c, simpleTestService, setInstances(instances))

	d1.GetService(c, id)
	d1.Stop(c)
	d1.Start(c)
	d1.GetService(c, id)

	d2 := s.AddDaemon(c, true, true)
	d2.GetService(c, id)
	d2.Stop(c)
	d2.Start(c)
	d2.GetService(c, id)

	d3 := s.AddDaemon(c, true, true)
	d3.GetService(c, id)
	d3.Stop(c)
	d3.Start(c)
	d3.GetService(c, id)

	d3.Kill()
	time.Sleep(1 * time.Second) // time to handle signal
	d3.Start(c)
	d3.GetService(c, id)
}

func (s *DockerSwarmSuite) TestAPISwarmScaleNoRollingUpdate(c *check.C) {
	d := s.AddDaemon(c, true, true)

	instances := 2
	id := d.CreateService(c, simpleTestService, setInstances(instances))

	waitAndAssert(c, defaultReconciliationTimeout, d.CheckActiveContainerCount, checker.Equals, instances)
	containers := d.ActiveContainers(c)
	instances = 4
	d.UpdateService(c, d.GetService(c, id), setInstances(instances))
	waitAndAssert(c, defaultReconciliationTimeout, d.CheckActiveContainerCount, checker.Equals, instances)
	containers2 := d.ActiveContainers(c)

loop0:
	for _, c1 := range containers {
		for _, c2 := range containers2 {
			if c1 == c2 {
				continue loop0
			}
		}
		c.Errorf("container %v not found in new set %#v", c1, containers2)
	}
}

func (s *DockerSwarmSuite) TestAPISwarmInvalidAddress(c *check.C) {
	d := s.AddDaemon(c, false, false)
	req := swarm.InitRequest{
		ListenAddr: "",
	}
	res, _, err := request.Post("/swarm/init", request.Host(d.Sock()), request.JSONBody(req))
	c.Assert(err, checker.IsNil)
	c.Assert(res.StatusCode, checker.Equals, http.StatusBadRequest)

	req2 := swarm.JoinRequest{
		ListenAddr:  "0.0.0.0:2377",
		RemoteAddrs: []string{""},
	}
	res, _, err = request.Post("/swarm/join", request.Host(d.Sock()), request.JSONBody(req2))
	c.Assert(err, checker.IsNil)
	c.Assert(res.StatusCode, checker.Equals, http.StatusBadRequest)
}

func (s *DockerSwarmSuite) TestAPISwarmForceNewCluster(c *check.C) {
	d1 := s.AddDaemon(c, true, true)
	d2 := s.AddDaemon(c, true, true)

	instances := 2
	id := d1.CreateService(c, simpleTestService, setInstances(instances))
	waitAndAssert(c, defaultReconciliationTimeout, reducedCheck(sumAsIntegers, d1.CheckActiveContainerCount, d2.CheckActiveContainerCount), checker.Equals, instances)

	// drain d2, all containers should move to d1
	d1.UpdateNode(c, d2.NodeID(), func(n *swarm.Node) {
		n.Spec.Availability = swarm.NodeAvailabilityDrain
	})
	waitAndAssert(c, defaultReconciliationTimeout, d1.CheckActiveContainerCount, checker.Equals, instances)
	waitAndAssert(c, defaultReconciliationTimeout, d2.CheckActiveContainerCount, checker.Equals, 0)

	d2.Stop(c)

	d1.SwarmInit(c, swarm.InitRequest{
		ForceNewCluster: true,
		Spec:            swarm.Spec{},
	})

	waitAndAssert(c, defaultReconciliationTimeout, d1.CheckActiveContainerCount, checker.Equals, instances)

	d3 := s.AddDaemon(c, true, true)
	info := d3.SwarmInfo(c)
	c.Assert(info.ControlAvailable, checker.True)
	c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)

	instances = 4
	d3.UpdateService(c, d3.GetService(c, id), setInstances(instances))

	waitAndAssert(c, defaultReconciliationTimeout, reducedCheck(sumAsIntegers, d1.CheckActiveContainerCount, d3.CheckActiveContainerCount), checker.Equals, instances)
}

func simpleTestService(s *swarm.Service) {
	ureplicas := uint64(1)
	restartDelay := time.Duration(100 * time.Millisecond)

	s.Spec = swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   "busybox:latest",
				Command: []string{"/bin/top"},
			},
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &ureplicas,
			},
		},
	}
	s.Spec.Name = "top"
}

func serviceForUpdate(s *swarm.Service) {
	ureplicas := uint64(1)
	restartDelay := time.Duration(100 * time.Millisecond)

	s.Spec = swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   "busybox:latest",
				Command: []string{"/bin/top"},
			},
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &ureplicas,
			},
		},
		UpdateConfig: &swarm.UpdateConfig{
			Parallelism:   2,
			Delay:         4 * time.Second,
			FailureAction: swarm.UpdateFailureActionContinue,
		},
		RollbackConfig: &swarm.UpdateConfig{
			Parallelism:   3,
			Delay:         4 * time.Second,
			FailureAction: swarm.UpdateFailureActionContinue,
		},
	}
	s.Spec.Name = "updatetest"
}

func setInstances(replicas int) testdaemon.ServiceConstructor {
	ureplicas := uint64(replicas)
	return func(s *swarm.Service) {
		s.Spec.Mode = swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &ureplicas,
			},
		}
	}
}

func setUpdateOrder(order string) testdaemon.ServiceConstructor {
	return func(s *swarm.Service) {
		if s.Spec.UpdateConfig == nil {
			s.Spec.UpdateConfig = &swarm.UpdateConfig{}
		}
		s.Spec.UpdateConfig.Order = order
	}
}

func setRollbackOrder(order string) testdaemon.ServiceConstructor {
	return func(s *swarm.Service) {
		if s.Spec.RollbackConfig == nil {
			s.Spec.RollbackConfig = &swarm.UpdateConfig{}
		}
		s.Spec.RollbackConfig.Order = order
	}
}

func setImage(image string) testdaemon.ServiceConstructor {
	return func(s *swarm.Service) {
		if s.Spec.TaskTemplate.ContainerSpec == nil {
			s.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
		}
		s.Spec.TaskTemplate.ContainerSpec.Image = image
	}
}

func setFailureAction(failureAction string) testdaemon.ServiceConstructor {
	return func(s *swarm.Service) {
		s.Spec.UpdateConfig.FailureAction = failureAction
	}
}

func setMaxFailureRatio(maxFailureRatio float32) testdaemon.ServiceConstructor {
	return func(s *swarm.Service) {
		s.Spec.UpdateConfig.MaxFailureRatio = maxFailureRatio
	}
}

func setParallelism(parallelism uint64) testdaemon.ServiceConstructor {
	return func(s *swarm.Service) {
		s.Spec.UpdateConfig.Parallelism = parallelism
	}
}

func setConstraints(constraints []string) testdaemon.ServiceConstructor {
	return func(s *swarm.Service) {
		if s.Spec.TaskTemplate.Placement == nil {
			s.Spec.TaskTemplate.Placement = &swarm.Placement{}
		}
		s.Spec.TaskTemplate.Placement.Constraints = constraints
	}
}

func setPlacementPrefs(prefs []swarm.PlacementPreference) testdaemon.ServiceConstructor {
	return func(s *swarm.Service) {
		if s.Spec.TaskTemplate.Placement == nil {
			s.Spec.TaskTemplate.Placement = &swarm.Placement{}
		}
		s.Spec.TaskTemplate.Placement.Preferences = prefs
	}
}

func setGlobalMode(s *swarm.Service) {
	s.Spec.Mode = swarm.ServiceMode{
		Global: &swarm.GlobalService{},
	}
}

func checkClusterHealth(c *check.C, cl []*daemon.Daemon, managerCount, workerCount int) {
	var totalMCount, totalWCount int

	for _, d := range cl {
		var (
			info swarm.Info
		)

		// check info in a waitAndAssert, because if the cluster doesn't have a leader, `info` will return an error
		checkInfo := func(c *check.C) (interface{}, check.CommentInterface) {
			client := d.NewClientT(c)
			daemonInfo, err := client.Info(context.Background())
			info = daemonInfo.Swarm
			return err, check.Commentf("cluster not ready in time")
		}
		waitAndAssert(c, defaultReconciliationTimeout, checkInfo, checker.IsNil)
		if !info.ControlAvailable {
			totalWCount++
			continue
		}

		var leaderFound bool
		totalMCount++
		var mCount, wCount int

		for _, n := range d.ListNodes(c) {
			waitReady := func(c *check.C) (interface{}, check.CommentInterface) {
				if n.Status.State == swarm.NodeStateReady {
					return true, nil
				}
				nn := d.GetNode(c, n.ID)
				n = *nn
				return n.Status.State == swarm.NodeStateReady, check.Commentf("state of node %s, reported by %s", n.ID, d.NodeID())
			}
			waitAndAssert(c, defaultReconciliationTimeout, waitReady, checker.True)

			waitActive := func(c *check.C) (interface{}, check.CommentInterface) {
				if n.Spec.Availability == swarm.NodeAvailabilityActive {
					return true, nil
				}
				nn := d.GetNode(c, n.ID)
				n = *nn
				return n.Spec.Availability == swarm.NodeAvailabilityActive, check.Commentf("availability of node %s, reported by %s", n.ID, d.NodeID())
			}
			waitAndAssert(c, defaultReconciliationTimeout, waitActive, checker.True)

			if n.Spec.Role == swarm.NodeRoleManager {
				c.Assert(n.ManagerStatus, checker.NotNil, check.Commentf("manager status of node %s (manager), reported by %s", n.ID, d.NodeID()))
				if n.ManagerStatus.Leader {
					leaderFound = true
				}
				mCount++
			} else {
				c.Assert(n.ManagerStatus, checker.IsNil, check.Commentf("manager status of node %s (worker), reported by %s", n.ID, d.NodeID()))
				wCount++
			}
		}
		c.Assert(leaderFound, checker.True, check.Commentf("lack of leader reported by node %s", info.NodeID))
		c.Assert(mCount, checker.Equals, managerCount, check.Commentf("managers count reported by node %s", info.NodeID))
		c.Assert(wCount, checker.Equals, workerCount, check.Commentf("workers count reported by node %s", info.NodeID))
	}
	c.Assert(totalMCount, checker.Equals, managerCount)
	c.Assert(totalWCount, checker.Equals, workerCount)
}

func (s *DockerSwarmSuite) TestAPISwarmRestartCluster(c *check.C) {
	mCount, wCount := 5, 1

	var nodes []*daemon.Daemon
	for i := 0; i < mCount; i++ {
		manager := s.AddDaemon(c, true, true)
		info := manager.SwarmInfo(c)
		c.Assert(info.ControlAvailable, checker.True)
		c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)
		nodes = append(nodes, manager)
	}

	for i := 0; i < wCount; i++ {
		worker := s.AddDaemon(c, true, false)
		info := worker.SwarmInfo(c)
		c.Assert(info.ControlAvailable, checker.False)
		c.Assert(info.LocalNodeState, checker.Equals, swarm.LocalNodeStateActive)
		nodes = append(nodes, worker)
	}

	// stop whole cluster
	{
		var wg sync.WaitGroup
		wg.Add(len(nodes))
		errs := make(chan error, len(nodes))

		for _, d := range nodes {
			go func(daemon *daemon.Daemon) {
				defer wg.Done()
				if err := daemon.StopWithError(); err != nil {
					errs <- err
				}
			}(d)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			c.Assert(err, check.IsNil)
		}
	}

	// start whole cluster
	{
		var wg sync.WaitGroup
		wg.Add(len(nodes))
		errs := make(chan error, len(nodes))

		for _, d := range nodes {
			go func(daemon *daemon.Daemon) {
				defer wg.Done()
				if err := daemon.StartWithError("--iptables=false"); err != nil {
					errs <- err
				}
			}(d)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			c.Assert(err, check.IsNil)
		}
	}

	checkClusterHealth(c, nodes, mCount, wCount)
}

func (s *DockerSwarmSuite) TestAPISwarmServicesUpdateWithName(c *check.C) {
	d := s.AddDaemon(c, true, true)

	instances := 2
	id := d.CreateService(c, simpleTestService, setInstances(instances))
	waitAndAssert(c, defaultReconciliationTimeout, d.CheckActiveContainerCount, checker.Equals, instances)

	service := d.GetService(c, id)
	instances = 5

	setInstances(instances)(service)
	cli, err := d.NewClient()
	c.Assert(err, checker.IsNil)
	defer cli.Close()
	_, err = cli.ServiceUpdate(context.Background(), service.Spec.Name, service.Version, service.Spec, types.ServiceUpdateOptions{})
	c.Assert(err, checker.IsNil)
	waitAndAssert(c, defaultReconciliationTimeout, d.CheckActiveContainerCount, checker.Equals, instances)
}

// Unlocking an unlocked swarm results in an error
func (s *DockerSwarmSuite) TestAPISwarmUnlockNotLocked(c *check.C) {
	d := s.AddDaemon(c, true, true)
	err := d.SwarmUnlock(swarm.UnlockRequest{UnlockKey: "wrong-key"})
	c.Assert(err, checker.NotNil)
	c.Assert(err.Error(), checker.Contains, "swarm is not locked")
}

// #29885
func (s *DockerSwarmSuite) TestAPISwarmErrorHandling(c *check.C) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", defaultSwarmPort))
	c.Assert(err, checker.IsNil)
	defer ln.Close()
	d := s.AddDaemon(c, false, false)
	client := d.NewClientT(c)
	_, err = client.SwarmInit(context.Background(), swarm.InitRequest{
		ListenAddr: d.SwarmListenAddr(),
	})
	c.Assert(err, checker.NotNil)
	c.Assert(err.Error(), checker.Contains, "address already in use")
}

// Test case for 30242, where duplicate networks, with different drivers `bridge` and `overlay`,
// caused both scopes to be `swarm` for `docker network inspect` and `docker network ls`.
// This test makes sure the fixes correctly output scopes instead.
func (s *DockerSwarmSuite) TestAPIDuplicateNetworks(c *check.C) {
	d := s.AddDaemon(c, true, true)
	cli, err := d.NewClient()
	c.Assert(err, checker.IsNil)
	defer cli.Close()

	name := "foo"
	networkCreate := types.NetworkCreate{
		CheckDuplicate: false,
	}

	networkCreate.Driver = "bridge"

	n1, err := cli.NetworkCreate(context.Background(), name, networkCreate)
	c.Assert(err, checker.IsNil)

	networkCreate.Driver = "overlay"

	n2, err := cli.NetworkCreate(context.Background(), name, networkCreate)
	c.Assert(err, checker.IsNil)

	r1, err := cli.NetworkInspect(context.Background(), n1.ID, types.NetworkInspectOptions{})
	c.Assert(err, checker.IsNil)
	c.Assert(r1.Scope, checker.Equals, "local")

	r2, err := cli.NetworkInspect(context.Background(), n2.ID, types.NetworkInspectOptions{})
	c.Assert(err, checker.IsNil)
	c.Assert(r2.Scope, checker.Equals, "swarm")
}

// Test case for 30178
func (s *DockerSwarmSuite) TestAPISwarmHealthcheckNone(c *check.C) {
	// Issue #36386 can be a independent one, which is worth further investigation.
	c.Skip("Root cause of Issue #36386 is needed")
	d := s.AddDaemon(c, true, true)

	out, err := d.Cmd("network", "create", "-d", "overlay", "lb")
	c.Assert(err, checker.IsNil, check.Commentf(out))

	instances := 1
	d.CreateService(c, simpleTestService, setInstances(instances), func(s *swarm.Service) {
		if s.Spec.TaskTemplate.ContainerSpec == nil {
			s.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
		}
		s.Spec.TaskTemplate.ContainerSpec.Healthcheck = &container.HealthConfig{}
		s.Spec.TaskTemplate.Networks = []swarm.NetworkAttachmentConfig{
			{Target: "lb"},
		}
	})

	waitAndAssert(c, defaultReconciliationTimeout, d.CheckActiveContainerCount, checker.Equals, instances)

	containers := d.ActiveContainers(c)

	out, err = d.Cmd("exec", containers[0], "ping", "-c1", "-W3", "top")
	c.Assert(err, checker.IsNil, check.Commentf(out))
}

func (s *DockerSwarmSuite) TestSwarmRepeatedRootRotation(c *check.C) {
	m := s.AddDaemon(c, true, true)
	w := s.AddDaemon(c, true, false)

	info := m.SwarmInfo(c)

	currentTrustRoot := info.Cluster.TLSInfo.TrustRoot

	// rotate multiple times
	for i := 0; i < 4; i++ {
		var err error
		var cert, key []byte
		if i%2 != 0 {
			cert, _, key, err = initca.New(&csr.CertificateRequest{
				CN:         "newRoot",
				KeyRequest: csr.NewBasicKeyRequest(),
				CA:         &csr.CAConfig{Expiry: ca.RootCAExpiration},
			})
			c.Assert(err, checker.IsNil)
		}
		expectedCert := string(cert)
		m.UpdateSwarm(c, func(s *swarm.Spec) {
			s.CAConfig.SigningCACert = expectedCert
			s.CAConfig.SigningCAKey = string(key)
			s.CAConfig.ForceRotate++
		})

		// poll to make sure update succeeds
		var clusterTLSInfo swarm.TLSInfo
		for j := 0; j < 18; j++ {
			info := m.SwarmInfo(c)

			// the desired CA cert and key is always redacted
			c.Assert(info.Cluster.Spec.CAConfig.SigningCAKey, checker.Equals, "")
			c.Assert(info.Cluster.Spec.CAConfig.SigningCACert, checker.Equals, "")

			clusterTLSInfo = info.Cluster.TLSInfo

			// if root rotation is done and the trust root has changed, we don't have to poll anymore
			if !info.Cluster.RootRotationInProgress && clusterTLSInfo.TrustRoot != currentTrustRoot {
				break
			}

			// root rotation not done
			time.Sleep(250 * time.Millisecond)
		}
		if cert != nil {
			c.Assert(clusterTLSInfo.TrustRoot, checker.Equals, expectedCert)
		}
		// could take another second or two for the nodes to trust the new roots after they've all gotten
		// new TLS certificates
		for j := 0; j < 18; j++ {
			mInfo := m.GetNode(c, m.NodeID()).Description.TLSInfo
			wInfo := m.GetNode(c, w.NodeID()).Description.TLSInfo

			if mInfo.TrustRoot == clusterTLSInfo.TrustRoot && wInfo.TrustRoot == clusterTLSInfo.TrustRoot {
				break
			}

			// nodes don't trust root certs yet
			time.Sleep(250 * time.Millisecond)
		}

		c.Assert(m.GetNode(c, m.NodeID()).Description.TLSInfo, checker.DeepEquals, clusterTLSInfo)
		c.Assert(m.GetNode(c, w.NodeID()).Description.TLSInfo, checker.DeepEquals, clusterTLSInfo)
		currentTrustRoot = clusterTLSInfo.TrustRoot
	}
}

func (s *DockerSwarmSuite) TestAPINetworkInspectWithScope(c *check.C) {
	d := s.AddDaemon(c, true, true)

	name := "test-scoped-network"
	ctx := context.Background()
	apiclient, err := d.NewClient()
	assert.NilError(c, err)

	resp, err := apiclient.NetworkCreate(ctx, name, types.NetworkCreate{Driver: "overlay"})
	assert.NilError(c, err)

	network, err := apiclient.NetworkInspect(ctx, name, types.NetworkInspectOptions{})
	assert.NilError(c, err)
	assert.Check(c, is.Equal("swarm", network.Scope))
	assert.Check(c, is.Equal(resp.ID, network.ID))

	_, err = apiclient.NetworkInspect(ctx, name, types.NetworkInspectOptions{Scope: "local"})
	assert.Check(c, client.IsErrNotFound(err))
}

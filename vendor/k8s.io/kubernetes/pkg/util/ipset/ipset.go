/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ipset

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/glog"
	utilexec "k8s.io/utils/exec"
)

// Interface is an injectable interface for running ipset commands.  Implementations must be goroutine-safe.
type Interface interface {
	// FlushSet deletes all entries from a named set.
	FlushSet(set string) error
	// DestroySet deletes a named set.
	DestroySet(set string) error
	// DestroyAllSets deletes all sets.
	DestroyAllSets() error
	// CreateSet creates a new set.  It will ignore error when the set already exists if ignoreExistErr=true.
	CreateSet(set *IPSet, ignoreExistErr bool) error
	// AddEntry adds a new entry to the named set.  It will ignore error when the entry already exists if ignoreExistErr=true.
	AddEntry(entry string, set *IPSet, ignoreExistErr bool) error
	// DelEntry deletes one entry from the named set
	DelEntry(entry string, set string) error
	// Test test if an entry exists in the named set
	TestEntry(entry string, set string) (bool, error)
	// ListEntries lists all the entries from a named set
	ListEntries(set string) ([]string, error)
	// ListSets list all set names from kernel
	ListSets() ([]string, error)
	// GetVersion returns the "X.Y" version string for ipset.
	GetVersion() (string, error)
}

// IPSetCmd represents the ipset util.  We use ipset command for ipset execute.
const IPSetCmd = "ipset"

// EntryMemberPattern is the regular expression pattern of ipset member list.
// The raw output of ipset command `ipset list {set}` is similar to,
//Name: foobar
//Type: hash:ip,port
//Revision: 2
//Header: family inet hashsize 1024 maxelem 65536
//Size in memory: 16592
//References: 0
//Members:
//192.168.1.2,tcp:8080
//192.168.1.1,udp:53
var EntryMemberPattern = "(?m)^(.*\n)*Members:\n"

// VersionPattern is the regular expression pattern of ipset version string.
// ipset version output is similar to "v6.10".
var VersionPattern = "v[0-9]+\\.[0-9]+"

// IPSet implements an Interface to an set.
type IPSet struct {
	// Name is the set name.
	Name string
	// SetType specifies the ipset type.
	SetType Type
	// HashFamily specifies the protocol family of the IP addresses to be stored in the set.
	// The default is inet, i.e IPv4.  If users want to use IPv6, they should specify inet6.
	HashFamily string
	// HashSize specifies the hash table size of ipset.
	HashSize int
	// MaxElem specifies the max element number of ipset.
	MaxElem int
	// PortRange specifies the port range of bitmap:port type ipset.
	PortRange string
	// TODO: add comment message for ipset
}

// Validate checks if a given ipset is valid or not.
func (set *IPSet) Validate() bool {
	// Check if protocol is valid for `HashIPPort`, `HashIPPortIP` and `HashIPPortNet` type set.
	if set.SetType == HashIPPort || set.SetType == HashIPPortIP || set.SetType == HashIPPortNet {
		if valid := validateHashFamily(set.HashFamily); !valid {
			return false
		}
	}
	// check set type
	if valid := validateIPSetType(set.SetType); !valid {
		return false
	}
	// check port range for bitmap type set
	if set.SetType == BitmapPort {
		if valid := validatePortRange(set.PortRange); !valid {
			return false
		}
	}
	// check hash size value of ipset
	if set.HashSize <= 0 {

		return false
	}
	// check max elem value of ipset
	if set.MaxElem <= 0 {
		glog.Errorf("Invalid maxelem value %d, should be >0", set.MaxElem)
		return false
	}

	return true
}

// Entry represents a ipset entry.
type Entry struct {
	// IP is the entry's IP.  The IP address protocol corresponds to the HashFamily of IPSet.
	// All entries' IP addresses in the same ip set has same the protocol, IPv4 or IPv6.
	IP string
	// Port is the entry's Port.
	Port int
	// Protocol is the entry's Protocol.  The protocols of entries in the same ip set are all
	// the same.  The accepted protocols are TCP and UDP.
	Protocol string
	// Net is the entry's IP network address.  Network address with zero prefix size can NOT
	// be stored.
	Net string
	// IP2 is the entry's second IP.  IP2 may not be empty for `hash:ip,port,ip` type ip set.
	IP2 string
	// SetType is the type of ipset where the entry exists.
	SetType Type
}

// Validate checks if a given ipset entry is valid or not.  The set parameter is the ipset that entry belongs to.
func (e *Entry) Validate(set *IPSet) bool {
	if e.Port < 0 {
		glog.Errorf("Entry %v port number %d should be >=0 for ipset %v", e, e.Port, set)
		return false
	}
	switch e.SetType {
	case HashIPPort:
		// set default protocol to tcp if empty
		if len(e.Protocol) == 0 {
			e.Protocol = ProtocolTCP
		}

		if net.ParseIP(e.IP) == nil {
			glog.Errorf("Error parsing entry %v ip address %v for ipset %v", e, e.IP, set)
			return false
		}

		if valid := validateProtocol(e.Protocol); !valid {
			return false
		}
	case HashIPPortIP:
		// set default protocol to tcp if empty
		if len(e.Protocol) == 0 {
			e.Protocol = ProtocolTCP
		}

		if net.ParseIP(e.IP) == nil {
			glog.Errorf("Error parsing entry %v ip address %v for ipset %v", e, e.IP, set)
			return false
		}

		if valid := validateProtocol(e.Protocol); !valid {
			return false
		}

		// IP2 can not be empty for `hash:ip,port,ip` type ip set
		if net.ParseIP(e.IP2) == nil {
			glog.Errorf("Error parsing entry %v second ip address %v for ipset %v", e, e.IP2, set)
			return false
		}
	case HashIPPortNet:
		// set default protocol to tcp if empty
		if len(e.Protocol) == 0 {
			e.Protocol = ProtocolTCP
		}

		if net.ParseIP(e.IP) == nil {
			glog.Errorf("Error parsing entry %v ip address %v for ipset %v", e, e.IP, set)
			return false
		}

		if valid := validateProtocol(e.Protocol); !valid {
			return false
		}

		// Net can not be empty for `hash:ip,port,net` type ip set
		if _, ipNet, _ := net.ParseCIDR(e.Net); ipNet == nil {
			glog.Errorf("Error parsing entry %v ip net %v for ipset %v", e, e.Net, set)
			return false
		}
	case BitmapPort:
		// check if port number satisfies its ipset's requirement of port range
		if set == nil {
			glog.Errorf("Unable to reference ip set where the entry %v exists", e)
			return false
		}
		begin, end, err := parsePortRange(set.PortRange)
		if err != nil {
			glog.Errorf("Failed to parse set %v port range %s for ipset %v, error: %v", set, set.PortRange, set, err)
			return false
		}
		if e.Port < begin || e.Port > end {
			glog.Errorf("Entry %v port number %d is not in the port range %s of its ipset %v", e, e.Port, set.PortRange, set)
			return false
		}
	}

	return true
}

// String returns the string format for ipset entry.
func (e *Entry) String() string {
	switch e.SetType {
	case HashIPPort:
		// Entry{192.168.1.1, udp, 53} -> 192.168.1.1,udp:53
		// Entry{192.168.1.2, tcp, 8080} -> 192.168.1.2,tcp:8080
		return fmt.Sprintf("%s,%s:%s", e.IP, e.Protocol, strconv.Itoa(e.Port))
	case HashIPPortIP:
		// Entry{192.168.1.1, udp, 53, 10.0.0.1} -> 192.168.1.1,udp:53,10.0.0.1
		// Entry{192.168.1.2, tcp, 8080, 192.168.1.2} -> 192.168.1.2,tcp:8080,192.168.1.2
		return fmt.Sprintf("%s,%s:%s,%s", e.IP, e.Protocol, strconv.Itoa(e.Port), e.IP2)
	case HashIPPortNet:
		// Entry{192.168.1.2, udp, 80, 10.0.1.0/24} -> 192.168.1.2,udp:80,10.0.1.0/24
		// Entry{192.168.2,25, tcp, 8080, 10.1.0.0/16} -> 192.168.2,25,tcp:8080,10.1.0.0/16
		return fmt.Sprintf("%s,%s:%s,%s", e.IP, e.Protocol, strconv.Itoa(e.Port), e.Net)
	case BitmapPort:
		// Entry{53} -> 53
		// Entry{8080} -> 8080
		return strconv.Itoa(e.Port)
	}
	return ""
}

type runner struct {
	exec utilexec.Interface
}

// New returns a new Interface which will exec ipset.
func New(exec utilexec.Interface) Interface {
	return &runner{
		exec: exec,
	}
}

// CreateSet creates a new set,  it will ignore error when the set already exists if ignoreExistErr=true.
func (runner *runner) CreateSet(set *IPSet, ignoreExistErr bool) error {
	// Setting default values if not present
	if set.HashSize == 0 {
		set.HashSize = 1024
	}
	if set.MaxElem == 0 {
		set.MaxElem = 65536
	}
	// Default protocol is IPv4
	if set.HashFamily == "" {
		set.HashFamily = ProtocolFamilyIPV4
	}
	// Default ipset type is "hash:ip,port"
	if len(set.SetType) == 0 {
		set.SetType = HashIPPort
	}
	if len(set.PortRange) == 0 {
		set.PortRange = DefaultPortRange
	}

	// Validate ipset before creating
	valid := set.Validate()
	if !valid {
		return fmt.Errorf("error creating ipset since it's invalid")
	}
	return runner.createSet(set, ignoreExistErr)
}

// If ignoreExistErr is set to true, then the -exist option of ipset will be specified, ipset ignores the error
// otherwise raised when the same set (setname and create parameters are identical) already exists.
func (runner *runner) createSet(set *IPSet, ignoreExistErr bool) error {
	args := []string{"create", set.Name, string(set.SetType)}
	if set.SetType == HashIPPortIP || set.SetType == HashIPPort {
		args = append(args,
			"family", set.HashFamily,
			"hashsize", strconv.Itoa(set.HashSize),
			"maxelem", strconv.Itoa(set.MaxElem),
		)
	}
	if set.SetType == BitmapPort {
		args = append(args, "range", set.PortRange)
	}
	if ignoreExistErr {
		args = append(args, "-exist")
	}
	if _, err := runner.exec.Command(IPSetCmd, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("error creating ipset %s, error: %v", set.Name, err)
	}
	return nil
}

// AddEntry adds a new entry to the named set.
// If the -exist option is specified, ipset ignores the error otherwise raised when
// the same set (setname and create parameters are identical) already exists.
func (runner *runner) AddEntry(entry string, set *IPSet, ignoreExistErr bool) error {
	args := []string{"add", set.Name, entry}
	if ignoreExistErr {
		args = append(args, "-exist")
	}
	if _, err := runner.exec.Command(IPSetCmd, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("error adding entry %s, error: %v", entry, err)
	}
	return nil
}

// DelEntry is used to delete the specified entry from the set.
func (runner *runner) DelEntry(entry string, set string) error {
	if _, err := runner.exec.Command(IPSetCmd, "del", set, entry).CombinedOutput(); err != nil {
		return fmt.Errorf("error deleting entry %s: from set: %s, error: %v", entry, set, err)
	}
	return nil
}

// TestEntry is used to check whether the specified entry is in the set or not.
func (runner *runner) TestEntry(entry string, set string) (bool, error) {
	if out, err := runner.exec.Command(IPSetCmd, "test", set, entry).CombinedOutput(); err == nil {
		reg, e := regexp.Compile("NOT")
		if e == nil && reg.MatchString(string(out)) {
			return false, nil
		} else if e == nil {
			return true, nil
		} else {
			return false, fmt.Errorf("error testing entry: %s, error: %v", entry, e)
		}
	} else {
		return false, fmt.Errorf("error testing entry %s: %v (%s)", entry, err, out)
	}
}

// FlushSet deletes all entries from a named set.
func (runner *runner) FlushSet(set string) error {
	if _, err := runner.exec.Command(IPSetCmd, "flush", set).CombinedOutput(); err != nil {
		return fmt.Errorf("error flushing set: %s, error: %v", set, err)
	}
	return nil
}

// DestroySet is used to destroy a named set.
func (runner *runner) DestroySet(set string) error {
	if out, err := runner.exec.Command(IPSetCmd, "destroy", set).CombinedOutput(); err != nil {
		return fmt.Errorf("error destroying set %s, error: %v(%s)", set, err, out)
	}
	return nil
}

// DestroyAllSets is used to destroy all sets.
func (runner *runner) DestroyAllSets() error {
	if _, err := runner.exec.Command(IPSetCmd, "destroy").CombinedOutput(); err != nil {
		return fmt.Errorf("error destroying all sets, error: %v", err)
	}
	return nil
}

// ListSets list all set names from kernel
func (runner *runner) ListSets() ([]string, error) {
	out, err := runner.exec.Command(IPSetCmd, "list", "-n").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing all sets, error: %v", err)
	}
	return strings.Split(string(out), "\n"), nil
}

// ListEntries lists all the entries from a named set.
func (runner *runner) ListEntries(set string) ([]string, error) {
	if len(set) == 0 {
		return nil, fmt.Errorf("set name can't be nil")
	}
	out, err := runner.exec.Command(IPSetCmd, "list", set).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing set: %s, error: %v", set, err)
	}
	memberMatcher := regexp.MustCompile(EntryMemberPattern)
	list := memberMatcher.ReplaceAllString(string(out[:]), "")
	strs := strings.Split(list, "\n")
	results := make([]string, 0)
	for i := range strs {
		if len(strs[i]) > 0 {
			results = append(results, strs[i])
		}
	}
	return results, nil
}

// GetVersion returns the version string.
func (runner *runner) GetVersion() (string, error) {
	return getIPSetVersionString(runner.exec)
}

// getIPSetVersionString runs "ipset --version" to get the version string
// in the form of "X.Y", i.e "6.19"
func getIPSetVersionString(exec utilexec.Interface) (string, error) {
	cmd := exec.Command(IPSetCmd, "--version")
	cmd.SetStdin(bytes.NewReader([]byte{}))
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	versionMatcher := regexp.MustCompile(VersionPattern)
	match := versionMatcher.FindStringSubmatch(string(bytes))
	if match == nil {
		return "", fmt.Errorf("no ipset version found in string: %s", bytes)
	}
	return match[0], nil
}

// checks if port range is valid. The begin port number is not necessarily less than
// end port number - ipset util can accept it.  It means both 1-100 and 100-1 are valid.
func validatePortRange(portRange string) bool {
	strs := strings.Split(portRange, "-")
	if len(strs) != 2 {
		glog.Errorf("port range should be in the format of `a-b`")
		return false
	}
	for i := range strs {
		num, err := strconv.Atoi(strs[i])
		if err != nil {
			glog.Errorf("Failed to parse %s, error: %v", strs[i], err)
			return false
		}
		if num < 0 {
			glog.Errorf("port number %d should be >=0", num)
			return false
		}
	}
	return true
}

// checks if the given ipset type is valid.
func validateIPSetType(set Type) bool {
	for _, valid := range ValidIPSetTypes {
		if set == valid {
			return true
		}
	}
	glog.Errorf("Currently supported ipset types are: %v, %s is not supported", ValidIPSetTypes, set)
	return false
}

// checks if given hash family is supported in ipset
func validateHashFamily(family string) bool {
	if family == ProtocolFamilyIPV4 || family == ProtocolFamilyIPV6 {
		return true
	}
	glog.Errorf("Currently supported ip set hash families are: [%s, %s], %s is not supported", ProtocolFamilyIPV4, ProtocolFamilyIPV6, family)
	return false
}

// IsNotFoundError returns true if the error indicates "not found".  It parses
// the error string looking for known values, which is imperfect but works in
// practice.
func IsNotFoundError(err error) bool {
	es := err.Error()
	if strings.Contains(es, "does not exist") {
		// set with the same name already exists
		// xref: https://github.com/Olipro/ipset/blob/master/lib/errcode.c#L32-L33
		return true
	}
	if strings.Contains(es, "element is missing") {
		// entry is missing from the set
		// xref: https://github.com/Olipro/ipset/blob/master/lib/parse.c#L1904
		// https://github.com/Olipro/ipset/blob/master/lib/parse.c#L1925
		return true
	}
	return false
}

// checks if given protocol is supported in entry
func validateProtocol(protocol string) bool {
	if protocol == ProtocolTCP || protocol == ProtocolUDP {
		return true
	}
	glog.Errorf("Invalid entry's protocol: %s, supported protocols are [%s, %s]", protocol, ProtocolTCP, ProtocolUDP)
	return false
}

// parsePortRange parse the begin and end port from a raw string(format: a-b).  beginPort <= endPort
// in the return value.
func parsePortRange(portRange string) (beginPort int, endPort int, err error) {
	if len(portRange) == 0 {
		portRange = DefaultPortRange
	}

	strs := strings.Split(portRange, "-")
	if len(strs) != 2 {
		// port number -1 indicates invalid
		return -1, -1, fmt.Errorf("port range should be in the format of `a-b`")
	}
	for i := range strs {
		num, err := strconv.Atoi(strs[i])
		if err != nil {
			// port number -1 indicates invalid
			return -1, -1, err
		}
		if num < 0 {
			// port number -1 indicates invalid
			return -1, -1, fmt.Errorf("port number %d should be >=0", num)
		}
		if i == 0 {
			beginPort = num
			continue
		}
		endPort = num
		// switch when first port number > second port number
		if beginPort > endPort {
			endPort = beginPort
			beginPort = num
		}
	}
	return beginPort, endPort, nil
}

var _ = Interface(&runner{})

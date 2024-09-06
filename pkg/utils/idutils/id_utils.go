/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package idutils

import (
	"errors"
	"net"

	"github.com/golang/example/stringutil"
	"github.com/sony/sonyflake"
	"github.com/speps/go-hashids"

	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

var sf *sonyflake.Sonyflake

func init() {
	var st sonyflake.Settings
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		sf = sonyflake.NewSonyflake(sonyflake.Settings{
			MachineID: lower16BitIP,
		})
	}
	if sf == nil {
		sf = sonyflake.NewSonyflake(sonyflake.Settings{
			MachineID: lower16BitIPv6,
		})
	}
}

func GetIntId() uint64 {
	if sf == nil {
		panic(errors.New("invalid snowflake instance"))
	}
	id, err := sf.NextID()
	if err != nil {
		panic(err)
	}
	return id
}

// format likes: B6BZVN3mOPvx
func GetUuid(prefix string) string {
	id := GetIntId()
	hd := hashids.NewData()
	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}
	i, err := h.Encode([]int{int(id)})
	if err != nil {
		panic(err)
	}

	return prefix + stringutils.Reverse(i)
}

const Alphabet36 = "abcdefghijklmnopqrstuvwxyz1234567890"

// format likes: 300m50zn91nwz5
func GetUuid36(prefix string) string {
	id := GetIntId()
	hd := hashids.NewData()
	hd.Alphabet = Alphabet36
	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}
	i, err := h.Encode([]int{int(id)})
	if err != nil {
		panic(err)
	}

	return prefix + stringutil.Reverse(i)
}

func lower16BitIP() (uint16, error) {
	ip, err := IPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}

func IPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if ip == nil {
			continue
		}
		return ip, nil

	}
	return nil, errors.New("no ip address")
}

func lower16BitIPv6() (uint16, error) {
	ip, err := IPv6()
	if err != nil {
		return 0, err
	}
	return uint16(ip[14])<<8 + uint16(ip[15]), nil
}
func IPv6() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}
		if ipnet.IP.To4() != nil {
			continue
		}
		ip := ipnet.IP.To16()
		if ip == nil {
			continue
		}
		return ip, nil

	}
	return nil, errors.New("no ip address")
}

package topdown

import (
	"fmt"
	"math/big"
	"net"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

func getNetFromOperand(v ast.Value) (*net.IPNet, error) {
	subnetStringA, err := builtins.StringOperand(v, 1)
	if err != nil {
		return nil, err
	}

	_, cidrnet, err := net.ParseCIDR(string(subnetStringA))
	if err != nil {
		return nil, err
	}

	return cidrnet, nil
}

func getLastIP(cidr *net.IPNet) (net.IP, error) {
	prefixLen, bits := cidr.Mask.Size()
	if prefixLen == 0 && bits == 0 {
		// non-standard mask, see https://golang.org/pkg/net/#IPMask.Size
		return nil, fmt.Errorf("CIDR mask is in non-standard format")
	}
	var lastIP []byte
	if prefixLen == bits {
		// Special case for single ip address ranges ex: 192.168.1.1/32
		// We can just use the starting IP as the last IP
		lastIP = cidr.IP
	} else {
		// Use big.Int's so we can handle ipv6 addresses
		firstIPInt := new(big.Int)
		firstIPInt.SetBytes(cidr.IP)
		hostLen := uint(bits) - uint(prefixLen)
		lastIPInt := big.NewInt(1)
		lastIPInt.Lsh(lastIPInt, hostLen)
		lastIPInt.Sub(lastIPInt, big.NewInt(1))
		lastIPInt.Or(lastIPInt, firstIPInt)

		ipBytes := lastIPInt.Bytes()
		lastIP = make([]byte, bits/8)

		// Pack our IP bytes into the end of the return array,
		// since big.Int.Bytes() removes front zero padding.
		for i := 1; i <= len(lastIPInt.Bytes()); i++ {
			lastIP[len(lastIP)-i] = ipBytes[len(ipBytes)-i]
		}
	}

	return lastIP, nil
}

func builtinNetCIDRIntersects(a, b ast.Value) (ast.Value, error) {
	cidrnetA, err := getNetFromOperand(a)
	if err != nil {
		return nil, err
	}

	cidrnetB, err := getNetFromOperand(b)
	if err != nil {
		return nil, err
	}

	// If either net contains the others starting IP they are overlapping
	cidrsOverlap := (cidrnetA.Contains(cidrnetB.IP) || cidrnetB.Contains(cidrnetA.IP))

	return ast.Boolean(cidrsOverlap), nil
}

func builtinNetCIDRContains(a, b ast.Value) (ast.Value, error) {
	cidrnetA, err := getNetFromOperand(a)
	if err != nil {
		return nil, err
	}

	// b could be either an IP addressor CIDR string, try to parse it as an IP first, fall back to CIDR
	bStr, err := builtins.StringOperand(b, 1)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(string(bStr))
	if ip != nil {
		return ast.Boolean(cidrnetA.Contains(ip)), nil
	}

	// It wasn't an IP, try and parse it as a CIDR
	cidrnetB, err := getNetFromOperand(b)
	if err != nil {
		return nil, fmt.Errorf("not a valid textual representation of an IP address or CIDR: %s", string(bStr))
	}

	// We can determine if cidr A contains cidr B iff A contains the starting address of B and the last address in B.
	cidrContained := false
	if cidrnetA.Contains(cidrnetB.IP) {
		// Only spend time calculating the last IP if the starting IP is already verified to be in cidr A
		lastIP, err := getLastIP(cidrnetB)
		if err != nil {
			return nil, err
		}
		cidrContained = cidrnetA.Contains(lastIP)
	}

	return ast.Boolean(cidrContained), nil
}

func builtinNetCIDRExpand(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	ip, ipNet, err := net.ParseCIDR(string(s))
	if err != nil {
		return err
	}

	result := ast.NewSet()

	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {

		if bctx.Cancel != nil && bctx.Cancel.Cancelled() {
			return &Error{
				Code:    CancelErr,
				Message: "net.cidr_expand: timed out before generating all IP addresses",
			}
		}

		result.Add(ast.StringTerm(ip.String()))
	}

	return iter(ast.NewTerm(result))
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func init() {
	RegisterFunctionalBuiltin2(ast.NetCIDROverlap.Name, builtinNetCIDRContains)
	RegisterFunctionalBuiltin2(ast.NetCIDRIntersects.Name, builtinNetCIDRIntersects)
	RegisterFunctionalBuiltin2(ast.NetCIDRContains.Name, builtinNetCIDRContains)
	RegisterBuiltinFunc(ast.NetCIDRExpand.Name, builtinNetCIDRExpand)
}

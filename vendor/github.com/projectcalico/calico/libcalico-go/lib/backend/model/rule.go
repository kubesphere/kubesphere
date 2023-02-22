// Copyright (c) 2016-2018 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"fmt"
	"strconv"
	"strings"

	apiv3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/projectcalico/api/pkg/lib/numorstring"

	"github.com/projectcalico/calico/libcalico-go/lib/net"
)

type Rule struct {
	Action string `json:"action,omitempty"`

	IPVersion *int `json:"ip_version,omitempty" validate:"omitempty,ipVersion"`

	Protocol    *numorstring.Protocol `json:"protocol,omitempty" validate:"omitempty"`
	NotProtocol *numorstring.Protocol `json:"!protocol,omitempty" validate:"omitempty"`

	// ICMP validation notes: 0 is a valid (common) ICMP type and code.  Type = 255 is not assigned
	// to any protocol and the Linux kernel doesn't support matching on it so we validate against
	// it.
	ICMPType    *int `json:"icmp_type,omitempty" validate:"omitempty,gte=0,lt=255"`
	ICMPCode    *int `json:"icmp_code,omitempty" validate:"omitempty,gte=0,lte=255"`
	NotICMPType *int `json:"!icmp_type,omitempty" validate:"omitempty,gte=0,lt=255"`
	NotICMPCode *int `json:"!icmp_code,omitempty" validate:"omitempty,gte=0,lte=255"`

	SrcTag              string             `json:"src_tag,omitempty" validate:"omitempty,tag"`
	SrcNet              *net.IPNet         `json:"src_net,omitempty" validate:"omitempty"`
	SrcNets             []*net.IPNet       `json:"src_nets,omitempty" validate:"omitempty"`
	SrcSelector         string             `json:"src_selector,omitempty" validate:"omitempty,selector"`
	SrcPorts            []numorstring.Port `json:"src_ports,omitempty" validate:"omitempty,dive"`
	SrcService          string             `json:"src_service,omitempty" validate:"omitempty"`
	SrcServiceNamespace string             `json:"src_service_ns,omitempty" validate:"omitempty"`
	DstTag              string             `json:"dst_tag,omitempty" validate:"omitempty,tag"`
	DstSelector         string             `json:"dst_selector,omitempty" validate:"omitempty,selector"`
	DstNet              *net.IPNet         `json:"dst_net,omitempty" validate:"omitempty"`
	DstNets             []*net.IPNet       `json:"dst_nets,omitempty" validate:"omitempty"`
	DstPorts            []numorstring.Port `json:"dst_ports,omitempty" validate:"omitempty,dive"`
	DstService          string             `json:"dst_service,omitempty" validate:"omitempty"`
	DstServiceNamespace string             `json:"dst_service_ns,omitempty" validate:"omitempty"`

	NotSrcTag      string             `json:"!src_tag,omitempty" validate:"omitempty,tag"`
	NotSrcNet      *net.IPNet         `json:"!src_net,omitempty" validate:"omitempty"`
	NotSrcNets     []*net.IPNet       `json:"!src_nets,omitempty" validate:"omitempty"`
	NotSrcSelector string             `json:"!src_selector,omitempty" validate:"omitempty,selector"`
	NotSrcPorts    []numorstring.Port `json:"!src_ports,omitempty" validate:"omitempty,dive"`
	NotDstTag      string             `json:"!dst_tag,omitempty" validate:"omitempty"`
	NotDstSelector string             `json:"!dst_selector,omitempty" validate:"omitempty,selector"`
	NotDstNet      *net.IPNet         `json:"!dst_net,omitempty" validate:"omitempty"`
	NotDstNets     []*net.IPNet       `json:"!dst_nets,omitempty" validate:"omitempty"`
	NotDstPorts    []numorstring.Port `json:"!dst_ports,omitempty" validate:"omitempty,dive"`

	// These fields allow us to pass through the raw match criteria from the V3 datamodel unmodified.
	// The selectors above are formed in the update processor layer by combining the original
	// selectors, namespace selectors and service account selectors into one.
	OriginalSrcSelector               string   `json:"orig_src_selector,omitempty" validate:"omitempty,selector"`
	OriginalSrcNamespaceSelector      string   `json:"orig_src_namespace_selector,omitempty" validate:"omitempty,selector"`
	OriginalDstSelector               string   `json:"orig_dst_selector,omitempty" validate:"omitempty,selector"`
	OriginalDstNamespaceSelector      string   `json:"orig_dst_namespace_selector,omitempty" validate:"omitempty,selector"`
	OriginalNotSrcSelector            string   `json:"!orig_src_selector,omitempty" validate:"omitempty,selector"`
	OriginalNotDstSelector            string   `json:"!orig_dst_selector,omitempty" validate:"omitempty,selector"`
	OriginalSrcServiceAccountNames    []string `json:"orig_src_service_acct_names,omitempty" validate:"omitempty"`
	OriginalSrcServiceAccountSelector string   `json:"orig_src_service_acct_selector,omitempty" validate:"omitempty,selector"`
	OriginalDstServiceAccountNames    []string `json:"orig_dst_service_acct_names,omitempty" validate:"omitempty"`
	OriginalDstServiceAccountSelector string   `json:"orig_dst_service_acct_selector,omitempty" validate:"omitempty,selector"`

	// These fields allow us to pass through application layer selectors from the V3 datamodel.
	HTTPMatch *HTTPMatch `json:"http,omitempty" validate:"omitempty"`

	LogPrefix string `json:"log_prefix,omitempty" validate:"omitempty"`

	Metadata *RuleMetadata `json:"metadata,omitempty" validate:"omitempty"`
}

type HTTPMatch struct {
	Methods []string         `json:"methods,omitempty" validate:"omitempty"`
	Paths   []apiv3.HTTPPath `json:"paths,omitempty" validate:"omitempty"`
}

type RuleMetadata struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

func combineNets(n *net.IPNet, nets []*net.IPNet) []*net.IPNet {
	if n == nil {
		return nets
	}
	if len(nets) == 0 {
		return []*net.IPNet{n}
	}
	var combination = make([]*net.IPNet, len(nets)+1)
	copy(combination, nets)
	combination[len(nets)] = n
	return combination
}

func (r Rule) AllSrcNets() []*net.IPNet {
	return combineNets(r.SrcNet, r.SrcNets)
}

func (r Rule) AllDstNets() []*net.IPNet {
	return combineNets(r.DstNet, r.DstNets)
}

func (r Rule) AllNotSrcNets() []*net.IPNet {
	return combineNets(r.NotSrcNet, r.NotSrcNets)
}

func (r Rule) AllNotDstNets() []*net.IPNet {
	return combineNets(r.NotDstNet, r.NotDstNets)
}

func joinNets(nets []*net.IPNet) string {
	parts := make([]string, len(nets))
	for i, n := range nets {
		parts[i] = n.String()
	}
	return strings.Join(parts, ",")
}

func (r Rule) String() string {
	parts := make([]string, 0)
	// Action.
	if r.Action != "" {
		parts = append(parts, r.Action)
	} else {
		parts = append(parts, "Allow")
	}

	// Global packet attributes that don't depend on direction.
	if r.Protocol != nil {
		parts = append(parts, r.Protocol.String())
	}
	if r.NotProtocol != nil {
		parts = append(parts, "!"+r.NotProtocol.String())
	}

	if r.ICMPType != nil {
		parts = append(parts, "type", strconv.Itoa(*r.ICMPType))
	}
	if r.ICMPCode != nil {
		parts = append(parts, "code", strconv.Itoa(*r.ICMPCode))
	}
	if r.NotICMPType != nil {
		parts = append(parts, "!type", strconv.Itoa(*r.NotICMPType))
	}
	if r.NotICMPCode != nil {
		parts = append(parts, "!code", strconv.Itoa(*r.NotICMPCode))
	}

	{
		// Source attributes.  New block ensures that fromParts goes out-of-scope before
		// we calculate toParts.  This prevents copy/paste errors.
		fromParts := make([]string, 0)
		if len(r.SrcPorts) > 0 {
			srcPorts := make([]string, len(r.SrcPorts))
			for ii, port := range r.SrcPorts {
				srcPorts[ii] = port.String()
			}
			fromParts = append(fromParts, "ports", strings.Join(srcPorts, ","))
		}
		if r.SrcTag != "" {
			fromParts = append(fromParts, "tag", r.SrcTag)
		}
		if r.SrcSelector != "" {
			fromParts = append(fromParts, "selector", fmt.Sprintf("%#v", r.SrcSelector))
		}
		srcNets := r.AllSrcNets()
		if len(srcNets) != 0 {
			fromParts = append(fromParts, "cidr", joinNets(srcNets))
		}
		if len(r.NotSrcPorts) > 0 {
			notSrcPorts := make([]string, len(r.NotSrcPorts))
			for ii, port := range r.NotSrcPorts {
				notSrcPorts[ii] = port.String()
			}
			fromParts = append(fromParts, "!ports", strings.Join(notSrcPorts, ","))
		}
		if r.NotSrcTag != "" {
			fromParts = append(fromParts, "!tag", r.NotSrcTag)
		}
		if r.NotSrcSelector != "" {
			fromParts = append(fromParts, "!selector", fmt.Sprintf("%#v", r.NotSrcSelector))
		}
		notSrcNets := r.AllNotSrcNets()
		if len(notSrcNets) != 0 {
			fromParts = append(fromParts, "!cidr", joinNets(notSrcNets))
		}

		if len(fromParts) > 0 {
			parts = append(parts, "from")
			parts = append(parts, fromParts...)
		}
	}

	{
		// Destination attributes.
		toParts := make([]string, 0)
		if len(r.DstPorts) > 0 {
			DstPorts := make([]string, len(r.DstPorts))
			for ii, port := range r.DstPorts {
				DstPorts[ii] = port.String()
			}
			toParts = append(toParts, "ports", strings.Join(DstPorts, ","))
		}
		if r.DstTag != "" {
			toParts = append(toParts, "tag", r.DstTag)
		}
		if r.DstSelector != "" {
			toParts = append(toParts, "selector", fmt.Sprintf("%#v", r.DstSelector))
		}
		dstNets := r.AllDstNets()
		if len(dstNets) != 0 {
			toParts = append(toParts, "cidr", joinNets(dstNets))
		}
		if len(r.NotDstPorts) > 0 {
			notDstPorts := make([]string, len(r.NotDstPorts))
			for ii, port := range r.NotDstPorts {
				notDstPorts[ii] = port.String()
			}
			toParts = append(toParts, "!ports", strings.Join(notDstPorts, ","))
		}
		if r.NotDstTag != "" {
			toParts = append(toParts, "!tag", r.NotDstTag)
		}
		if r.NotDstSelector != "" {
			toParts = append(toParts, "!selector", fmt.Sprintf("%#v", r.NotDstSelector))
		}
		notDstNets := r.AllNotDstNets()
		if len(notDstNets) != 0 {
			toParts = append(toParts, "!cidr", joinNets(notDstNets))
		}

		// HTTPMatch are destination rules.
		if r.HTTPMatch != nil {
			if len(r.HTTPMatch.Methods) > 0 {
				toParts = append(toParts, "httpMethods", fmt.Sprintf("%+v", r.HTTPMatch.Methods))
			}
			if len(r.HTTPMatch.Paths) > 0 {
				toParts = append(toParts, "httpPaths", fmt.Sprintf("%+v", r.HTTPMatch.Paths))
			}
		}

		if len(toParts) > 0 {
			parts = append(parts, "to")
			parts = append(parts, toParts...)
		}
	}

	return strings.Join(parts, " ")
}

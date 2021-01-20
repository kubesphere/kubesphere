package v1

import (
	"encoding/json"
	"fmt"

	"k8s.io/klog"
)

func (ss *SubnetStatus) Bytes() ([]byte, error) {
	//{"availableIPs":65527,"usingIPs":9} => {"status": {"availableIPs":65527,"usingIPs":9}}
	bytes, err := json.Marshal(ss)
	if err != nil {
		return nil, err
	}
	newStr := fmt.Sprintf(`{"status": %s}`, string(bytes))
	klog.V(5).Info("status body", newStr)
	return []byte(newStr), nil
}

func (vs *VlanStatus) Bytes() ([]byte, error) {
	bytes, err := json.Marshal(vs)
	if err != nil {
		return nil, err
	}
	newStr := fmt.Sprintf(`{"status": %s}`, string(bytes))
	klog.V(5).Info("status body", newStr)
	return []byte(newStr), nil
}

func (vs *VpcStatus) Bytes() ([]byte, error) {
	bytes, err := json.Marshal(vs)
	if err != nil {
		return nil, err
	}
	newStr := fmt.Sprintf(`{"status": %s}`, string(bytes))
	klog.V(5).Info("status body", newStr)
	return []byte(newStr), nil
}

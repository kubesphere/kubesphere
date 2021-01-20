package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *SubnetStatus) addCondition(ctype ConditionType, status corev1.ConditionStatus, reason, message string) {
	now := metav1.Now()
	c := &SubnetCondition{
		Type:               ctype,
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Status:             status,
		Reason:             reason,
		Message:            message,
	}
	m.Conditions = append(m.Conditions, *c)
}

// setConditionValue updates or creates a new condition
func (m *SubnetStatus) setConditionValue(ctype ConditionType, status corev1.ConditionStatus, reason, message string) {
	var c *SubnetCondition
	for i := range m.Conditions {
		if m.Conditions[i].Type == ctype {
			c = &m.Conditions[i]
		}
	}
	if c == nil {
		m.addCondition(ctype, status, reason, message)
	} else {
		// check message ?
		if c.Status == status && c.Reason == reason && c.Message == message {
			return
		}
		now := metav1.Now()
		c.LastUpdateTime = now
		if c.Status != status {
			c.LastTransitionTime = now
		}
		c.Status = status
		c.Reason = reason
		c.Message = message
	}
}

// RemoveCondition removes the condition with the provided type.
func (m *SubnetStatus) RemoveCondition(ctype ConditionType) {
	for i, c := range m.Conditions {
		if c.Type == ctype {
			m.Conditions[i] = m.Conditions[len(m.Conditions)-1]
			m.Conditions = m.Conditions[:len(m.Conditions)-1]
			break
		}
	}
}

// GetCondition get existing condition
func (m *SubnetStatus) GetCondition(ctype ConditionType) *SubnetCondition {
	for i := range m.Conditions {
		if m.Conditions[i].Type == ctype {
			return &m.Conditions[i]
		}
	}
	return nil
}

// IsConditionTrue - if condition is true
func (m *SubnetStatus) IsConditionTrue(ctype ConditionType) bool {
	if c := m.GetCondition(ctype); c != nil {
		return c.Status == corev1.ConditionTrue
	}
	return false
}

// IsReady returns true if ready condition is set
func (m *SubnetStatus) IsReady() bool { return m.IsConditionTrue(Ready) }

// IsNotReady returns true if ready condition is set
func (m *SubnetStatus) IsNotReady() bool { return !m.IsConditionTrue(Ready) }

// IsValidated returns true if ready condition is set
func (m *SubnetStatus) IsValidated() bool { return m.IsConditionTrue(Validated) }

// IsNotValidated returns true if ready condition is set
func (m *SubnetStatus) IsNotValidated() bool { return !m.IsConditionTrue(Validated) }

// ConditionReason - return condition reason
func (m *SubnetStatus) ConditionReason(ctype ConditionType) string {
	if c := m.GetCondition(ctype); c != nil {
		return c.Reason
	}
	return ""
}

// Ready - shortcut to set ready condition to true
func (m *SubnetStatus) Ready(reason, message string) {
	m.SetCondition(Ready, reason, message)
}

// NotReady - shortcut to set ready condition to false
func (m *SubnetStatus) NotReady(reason, message string) {
	m.ClearCondition(Ready, reason, message)
}

// Validated - shortcut to set validated condition to true
func (m *SubnetStatus) Validated(reason, message string) {
	m.SetCondition(Validated, reason, message)
}

// NotValidated - shortcut to set validated condition to false
func (m *SubnetStatus) NotValidated(reason, message string) {
	m.ClearCondition(Validated, reason, message)
}

// SetError - shortcut to set error condition
func (m *SubnetStatus) SetError(reason, message string) {
	m.SetCondition(Error, reason, message)
}

// ClearError - shortcut to set error condition
func (m *SubnetStatus) ClearError() {
	m.ClearCondition(Error, "NoError", "No error seen")
}

// EnsureCondition useful for adding default conditions
func (m *SubnetStatus) EnsureCondition(ctype ConditionType) {
	if c := m.GetCondition(ctype); c != nil {
		return
	}
	m.addCondition(ctype, corev1.ConditionUnknown, ReasonInit, "Not Observed")
}

// EnsureStandardConditions - helper to inject standard conditions
func (m *SubnetStatus) EnsureStandardConditions() {
	m.EnsureCondition(Ready)
	m.EnsureCondition(Validated)
	m.EnsureCondition(Error)
}

// ClearCondition updates or creates a new condition
func (m *SubnetStatus) ClearCondition(ctype ConditionType, reason, message string) {
	m.setConditionValue(ctype, corev1.ConditionFalse, reason, message)
}

// SetCondition updates or creates a new condition
func (m *SubnetStatus) SetCondition(ctype ConditionType, reason, message string) {
	m.setConditionValue(ctype, corev1.ConditionTrue, reason, message)
}

// RemoveAllConditions updates or creates a new condition
func (m *SubnetStatus) RemoveAllConditions() {
	m.Conditions = []SubnetCondition{}
}

// ClearAllConditions updates or creates a new condition
func (m *SubnetStatus) ClearAllConditions() {
	for i := range m.Conditions {
		m.Conditions[i].Status = corev1.ConditionFalse
	}
}

// SetError - shortcut to set error condition
func (v *VlanStatus) SetVlanError(reason, message string) {
	v.SetVlanCondition(Error, reason, message)
}

// SetCondition updates or creates a new condition
func (v *VlanStatus) SetVlanCondition(ctype ConditionType, reason, message string) {
	v.setVlanConditionValue(ctype, corev1.ConditionTrue, reason, message)
}

func (v *VlanStatus) setVlanConditionValue(ctype ConditionType, status corev1.ConditionStatus, reason, message string) {
	var c *VlanCondition
	for i := range v.Conditions {
		if v.Conditions[i].Type == ctype {
			c = &v.Conditions[i]
		}
	}
	if c == nil {
		v.addVlanCondition(ctype, status, reason, message)
	} else {
		// check message ?
		if c.Status == status && c.Reason == reason && c.Message == message {
			return
		}
		now := metav1.Now()
		c.LastUpdateTime = now
		if c.Status != status {
			c.LastTransitionTime = now
		}
		c.Status = status
		c.Reason = reason
		c.Message = message
	}
}

func (v *VlanStatus) addVlanCondition(ctype ConditionType, status corev1.ConditionStatus, reason, message string) {
	now := metav1.Now()
	c := &VlanCondition{
		Type:               ctype,
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Status:             status,
		Reason:             reason,
		Message:            message,
	}
	v.Conditions = append(v.Conditions, *c)
}

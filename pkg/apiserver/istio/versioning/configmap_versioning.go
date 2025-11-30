package versioning

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ConfigMapVersionManager manages ConfigMap versions
type ConfigMapVersionManager struct {
	k8sClient kubernetes.Interface
}

// NewVersionManager creates a new ConfigMapVersionManager
func NewVersionManager(k8sClient kubernetes.Interface) *ConfigMapVersionManager {
	return &ConfigMapVersionManager{
		k8sClient: k8sClient,
	}
}

// CreateVersion creates a new version of a ConfigMap
func (m *ConfigMapVersionManager) CreateVersion(
	ctx context.Context,
	configMapName string,
) (string, error) {
	return m.CreateVersionInNamespace(ctx, configMapName, metav1.NamespaceDefault)
}

// CreateVersionInNamespace creates a new version of a ConfigMap in a specific namespace
func (m *ConfigMapVersionManager) CreateVersionInNamespace(
	ctx context.Context,
	configMapName, namespace string,
) (string, error) {
	// Get original ConfigMap
	cm, err := m.k8sClient.CoreV1().ConfigMaps(namespace).Get(
		ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get ConfigMap: %v", err)
	}

	// Generate version hash
	version := m.generateVersionHash(cm)

	// Create versioned ConfigMap
	versionedCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", configMapName, version),
			Namespace: cm.Namespace,
			Labels: map[string]string{
				"configmap-version":   version,
				"original-configmap":  configMapName,
				"managed-by":          "kubesphere-grayscale",
				"version-timestamp":   time.Now().Format("20060102-150405"),
			},
			Annotations: map[string]string{
				"version-timestamp": metav1.Now().Format(time.RFC3339),
				"version-hash":      version,
				"original-name":     configMapName,
				"created-by":        "configmap-grayscale-manager",
			},
		},
		Data:       cm.Data,
		BinaryData: cm.BinaryData,
	}

	_, err = m.k8sClient.CoreV1().ConfigMaps(cm.Namespace).Create(
		ctx, versionedCM, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create versioned ConfigMap: %v", err)
	}

	return version, nil
}

// ListVersions lists all versions of a ConfigMap
func (m *ConfigMapVersionManager) ListVersions(
	ctx context.Context,
	configMapName string,
) ([]*ConfigMapVersion, error) {
	return m.ListVersionsInNamespace(ctx, configMapName, metav1.NamespaceDefault)
}

// ListVersionsInNamespace lists all versions of a ConfigMap in a specific namespace
func (m *ConfigMapVersionManager) ListVersionsInNamespace(
	ctx context.Context,
	configMapName, namespace string,
) ([]*ConfigMapVersion, error) {
	// List ConfigMaps with version label
	labelSelector := fmt.Sprintf("original-configmap=%s,managed-by=kubesphere-grayscale", configMapName)
	configMaps, err := m.k8sClient.CoreV1().ConfigMaps(namespace).List(
		ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("failed to list ConfigMap versions: %v", err)
	}

	var versions []*ConfigMapVersion
	for _, cm := range configMaps.Items {
		version := &ConfigMapVersion{
			Name:        cm.Name,
			Version:     cm.Labels["configmap-version"],
			CreatedAt:   cm.CreationTimestamp,
			Size:        m.calculateConfigMapSize(&cm),
			DataKeys:    m.getDataKeys(&cm),
			BinaryKeys:  m.getBinaryKeys(&cm),
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// GetVersion gets a specific version of a ConfigMap
func (m *ConfigMapVersionManager) GetVersion(
	ctx context.Context,
	configMapName, version string,
) (*corev1.ConfigMap, error) {
	return m.GetVersionInNamespace(ctx, configMapName, version, metav1.NamespaceDefault)
}

// GetVersionInNamespace gets a specific version of a ConfigMap in a specific namespace
func (m *ConfigMapVersionManager) GetVersionInNamespace(
	ctx context.Context,
	configMapName, version, namespace string,
) (*corev1.ConfigMap, error) {
	versionedName := fmt.Sprintf("%s-%s", configMapName, version)
	cm, err := m.k8sClient.CoreV1().ConfigMaps(namespace).Get(
		ctx, versionedName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ConfigMap version: %v", err)
	}

	return cm, nil
}

// DeleteVersion deletes a specific version of a ConfigMap
func (m *ConfigMapVersionManager) DeleteVersion(
	ctx context.Context,
	configMapName, version string,
) error {
	return m.DeleteVersionInNamespace(ctx, configMapName, version, metav1.NamespaceDefault)
}

// DeleteVersionInNamespace deletes a specific version of a ConfigMap in a specific namespace
func (m *ConfigMapVersionManager) DeleteVersionInNamespace(
	ctx context.Context,
	configMapName, version, namespace string,
) error {
	versionedName := fmt.Sprintf("%s-%s", configMapName, version)
	err := m.k8sClient.CoreV1().ConfigMaps(namespace).Delete(
		ctx, versionedName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete ConfigMap version: %v", err)
	}

	return nil
}

// CleanupOldVersions removes old versions of a ConfigMap, keeping only the latest N versions
func (m *ConfigMapVersionManager) CleanupOldVersions(
	ctx context.Context,
	configMapName string,
	keepCount int,
) error {
	return m.CleanupOldVersionsInNamespace(ctx, configMapName, keepCount, metav1.NamespaceDefault)
}

// CleanupOldVersionsInNamespace removes old versions of a ConfigMap in a specific namespace
func (m *ConfigMapVersionManager) CleanupOldVersionsInNamespace(
	ctx context.Context,
	configMapName string,
	keepCount int,
	namespace string,
) error {
	versions, err := m.ListVersionsInNamespace(ctx, configMapName, namespace)
	if err != nil {
		return fmt.Errorf("failed to list versions: %v", err)
	}

	if len(versions) <= keepCount {
		return nil // No cleanup needed
	}

	// Sort versions by creation time (oldest first)
	// Sort versions by creation time (newest first)
	for i := 0; i < len(versions)-1; i++ {
		for j := i + 1; j < len(versions); j++ {
			if versions[i].CreatedAt.Before(&versions[j].CreatedAt) {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}

	// Delete old versions (keep the latest N)
	for i := keepCount; i < len(versions); i++ {
		if err := m.DeleteVersionInNamespace(ctx, configMapName, versions[i].Version, namespace); err != nil {
			return fmt.Errorf("failed to delete version %s: %v", versions[i].Version, err)
		}
	}

	return nil
}

// CompareVersions compares two versions of a ConfigMap
func (m *ConfigMapVersionManager) CompareVersions(
	ctx context.Context,
	configMapName, version1, version2 string,
) (*ConfigMapDiff, error) {
	return m.CompareVersionsInNamespace(ctx, configMapName, version1, version2, metav1.NamespaceDefault)
}

// CompareVersionsInNamespace compares two versions of a ConfigMap in a specific namespace
func (m *ConfigMapVersionManager) CompareVersionsInNamespace(
	ctx context.Context,
	configMapName, version1, version2, namespace string,
) (*ConfigMapDiff, error) {
	cm1, err := m.GetVersionInNamespace(ctx, configMapName, version1, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %s: %v", version1, err)
	}

	cm2, err := m.GetVersionInNamespace(ctx, configMapName, version2, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %s: %v", version2, err)
	}

	diff := &ConfigMapDiff{
		Version1: version1,
		Version2: version2,
		Added:    []string{},
		Modified: []ConfigMapKeyDiff{},
		Deleted:  []string{},
	}

	// Compare data keys
	allKeys := make(map[string]bool)
	for key := range cm1.Data {
		allKeys[key] = true
	}
	for key := range cm2.Data {
		allKeys[key] = true
	}

	for key := range allKeys {
		val1, exists1 := cm1.Data[key]
		val2, exists2 := cm2.Data[key]

		if !exists1 && exists2 {
			diff.Added = append(diff.Added, key)
		} else if exists1 && !exists2 {
			diff.Deleted = append(diff.Deleted, key)
		} else if exists1 && exists2 && val1 != val2 {
			diff.Modified = append(diff.Modified, ConfigMapKeyDiff{
				Key:     key,
				OldValue: val1,
				NewValue: val2,
			})
		}
	}

	return diff, nil
}

// generateVersionHash generates a hash for the ConfigMap content
func (m *ConfigMapVersionManager) generateVersionHash(cm *corev1.ConfigMap) string {
	hasher := sha256.New()

	// Hash all data
	for key, value := range cm.Data {
		hasher.Write([]byte(key))
		hasher.Write([]byte(value))
	}

	// Hash binary data
	for key, value := range cm.BinaryData {
		hasher.Write([]byte(key))
		hasher.Write(value)
	}

	// Hash metadata for uniqueness
	hasher.Write([]byte(cm.Namespace))
	hasher.Write([]byte(cm.Name))

	return hex.EncodeToString(hasher.Sum(nil))[:12]
}

// calculateConfigMapSize calculates the size of a ConfigMap
func (m *ConfigMapVersionManager) calculateConfigMapSize(cm *corev1.ConfigMap) int {
	size := 0
	for _, value := range cm.Data {
		size += len(value)
	}
	for _, value := range cm.BinaryData {
		size += len(value)
	}
	return size
}

// getDataKeys returns the data keys of a ConfigMap
func (m *ConfigMapVersionManager) getDataKeys(cm *corev1.ConfigMap) []string {
	keys := make([]string, 0, len(cm.Data))
	for key := range cm.Data {
		keys = append(keys, key)
	}
	return keys
}

// getBinaryKeys returns the binary data keys of a ConfigMap
func (m *ConfigMapVersionManager) getBinaryKeys(cm *corev1.ConfigMap) []string {
	keys := make([]string, 0, len(cm.BinaryData))
	for key := range cm.BinaryData {
		keys = append(keys, key)
	}
	return keys
}

// ConfigMapVersion represents a versioned ConfigMap
type ConfigMapVersion struct {
	Name      string
	Version   string
	CreatedAt metav1.Time
	Size      int
	DataKeys  []string
	BinaryKeys []string
}

// ConfigMapDiff represents the difference between two ConfigMap versions
type ConfigMapDiff struct {
	Version1 string
	Version2 string
	Added    []string
	Modified []ConfigMapKeyDiff
	Deleted  []string
}

// ConfigMapKeyDiff represents a difference in a specific ConfigMap key
type ConfigMapKeyDiff struct {
	Key      string
	OldValue string
	NewValue string
}

package v1alpha1

const (
	StateUnknown         = ""
	StateInstalling      = "Installing"
	StateInstallFailed   = "InstallFailed"
	StateUpgrading       = "Upgrading"
	StateUpgradeFailed   = "UpgradeFailed"
	StateInstalled       = "Installed"
	StateDeployed        = StateInstalled
	StateUpgraded        = "Upgraded"
	StateUninstalling    = "Uninstalling"
	StateUninstalled     = "Uninstalled"
	StateUninstallFailed = "UninstallFailed"
	// StatePreparing indicates that the Extension is in the Preparing state.
	// This value is only used for Extension objects and is triggered when the state of its InstallPlan is empty
	// and is changing to the Installing/Upgrading state.
	StatePreparing = "Preparing"

	MaxStateNum = 10

	ConditionTypeInitialized = "Initialized"
	ConditionTypeInstalled   = "Installed"
	ConditionTypeUpgraded    = "Upgraded"
	ConditionTypeUninstalled = "Uninstalled"
	ConditionTypeReady       = "Ready"

	DisplayNameAnnotation          = "kubesphere.io/display-name"
	KSVersionAnnotation            = "kubesphere.io/ks-version"
	InstallationModeAnnotation     = "kubesphere.io/installation-mode"
	ExternalDependenciesAnnotation = "kubesphere.io/external-dependencies"

	ExtensionReferenceLabel  = "kubesphere.io/extension-ref"
	RepositoryReferenceLabel = "kubesphere.io/repository-ref"
	CategoryLabel            = "kubesphere.io/category"

	ForceDeleteAnnotation                = "kubesphere.io/force-delete"
	ExecutorHookImageAnnotation          = "executor-hook-image.kubesphere.io"
	ExecutorInstallHookImageAnnotation   = "executor-hook-image.kubesphere.io/install"
	ExecutorUpgradeHookImageAnnotation   = "executor-hook-image.kubesphere.io/upgrade"
	ExecutorUninstallHookImageAnnotation = "executor-hook-image.kubesphere.io/uninstall"
)

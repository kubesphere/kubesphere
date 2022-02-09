package runners

// AutoResizeEnabledKey is the key of flag that enables pvc-autoresizer.
const AutoResizeEnabledKey = "resize.kubesphere.io/enabled"

// ResizeThresholdAnnotation is the key of resize threshold.
const ResizeThresholdAnnotation = "resize.kubesphere.io/threshold"

// ResizeInodesThresholdAnnotation is the key of resize threshold for inodes.
const ResizeInodesThresholdAnnotation = "resize.kubesphere.io/inodes-threshold"

// ResizeIncreaseAnnotation is the key of amount increased.
const ResizeIncreaseAnnotation = "resize.kubesphere.io/increase"

// StorageLimitAnnotation is the key of storage limit value
const StorageLimitAnnotation = "resize.kubesphere.io/storage-limit"

// PreviousCapacityBytesAnnotation is the key of previous volume capacity.
const PreviousCapacityBytesAnnotation = "resize.kubesphere.io/pre-capacity-bytes"

// AutoRestartEnabledKey  is the key of flag that enables pods-autoRestart.
const AutoRestartEnabledKey = "restart.kubesphere.io/enabled"

// SupportOnlineResize is the key of flag that the storage class support online expansion
const SupportOnlineResize = "restart.kubesphere.io/online-expansion-support"

// RestartSkip is the key of flag that the workload don't need autoRestart
const RestartSkip = "restart.kubesphere.io/skip"

// ResizingMaxTime is the key of flag that the maximum number of seconds that autoRestart can wait for pvc resize
const ResizingMaxTime = "restart.kubesphere.io/max-time"

// RestartStage is used to record whether autoRestart has finished shutting down the pod
const RestartStage = "restart.kubesphere.io/stage"

// RestartStopTime is used to record the time when the pod is closed
const RestartStopTime = "restart.kubesphere.io/stop-time"

// ExpectReplicaNums is used to record the value of replicas before restart
const ExpectReplicaNums = "restart.kubesphere.io/replica-nums"

// DefaultThreshold is the default value of ResizeThresholdAnnotation.
const DefaultThreshold = "10%"

// DefaultInodesThreshold is the default value of ResizeInodesThresholdAnnotation.
const DefaultInodesThreshold = "10%"

// DefaultIncrease is the default value of ResizeIncreaseAnnotation.
const DefaultIncrease = "10%"

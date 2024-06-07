package rbac

import "fmt"

const iamPrefix = "kubesphere:iam"

func RelatedK8sResourceName(name string) string {
	return fmt.Sprintf("%s:%s", iamPrefix, name)
}

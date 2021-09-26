package v1beta1

import (
	ctschema "github.com/open-policy-agent/frameworks/constraint/pkg/schema"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/defaulting"
	"k8s.io/apimachinery/pkg/runtime"
)

const version = "v1beta1"

var (
	structuralSchema *schema.Structural
	Scheme           *runtime.Scheme
)

func init() {
	Scheme = runtime.NewScheme()
	var err error
	if err = apiextensionsv1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err = apiextensions.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err = AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if structuralSchema, err = ctschema.CRDSchema(Scheme, version); err != nil {
		panic(err)
	}
}

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_ConstraintTemplate(obj *ConstraintTemplate) { // nolint:revive // Required exact function name.
	// turn the CT into an unstructured
	un, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		panic("Failed to convert v1 ConstraintTemplate to Unstructured")
	}

	defaulting.Default(un, structuralSchema)

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(un, obj)
	if err != nil {
		panic("Failed to convert Unstructured to v1 ConstraintTemplate")
	}
}

/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package conversion

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook/conversion"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const webhookName = "conversion-webhook"

func (w *Webhook) Name() string {
	return webhookName
}

var _ kscontroller.Controller = &Webhook{}

type Webhook struct {
}

func (w *Webhook) SetupWithManager(mgr *kscontroller.Manager) error {
	mgr.GetWebhookServer().Register("/convert", conversion.NewWebhookHandler(mgr.GetScheme()))
	return nil
}

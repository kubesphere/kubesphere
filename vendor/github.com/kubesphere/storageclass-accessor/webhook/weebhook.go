package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var (
	certFile string
	keyFile  string
	port     int
)

// CmdWebhook is user by Cobra
var CmdWebhook = &cobra.Command{
	Use:  "storageClass-webhook",
	Args: cobra.MaximumNArgs(0),
	Run:  main,
}

func init() {
	CmdWebhook.Flags().StringVar(&certFile, "tls-cert-file", "",
		"File containing the x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert). Required.")
	CmdWebhook.Flags().StringVar(&keyFile, "tls-private-key-file", "",
		"File containing the x509 private key matching --tls-cert-file. Required.")
	CmdWebhook.Flags().IntVar(&port, "port", 443,
		"Secure port that the webhook listens on")
	CmdWebhook.MarkFlagRequired("tls-cert-file")
	CmdWebhook.MarkFlagRequired("tls-private-key-file")
}

// admitV1beta1Func handles a v1 admission
type admitV1Func func(v1.AdmissionReview) *v1.AdmissionResponse

// admitHandler is a handler, for both validators and mutators, that supports multiple admission review versions
type admitHandler struct {
	v1 admitV1Func
}

func newDelegateToV1AdmitHandler(f admitV1Func) admitHandler {
	return admitHandler{v1: f}
}

func server(w http.ResponseWriter, r *http.Request, admit admitHandler) {
	var err error
	if r.Body == nil {
		err = fmt.Errorf("request body is nil")
		klog.ErrorS(err, "request body is nil")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var body []byte
	body, err = io.ReadAll(r.Body)
	if err != nil {
		klog.ErrorS(err, "read request body failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		err = fmt.Errorf("contentType=%s, expect application/json", contentType)
		klog.ErrorS(err, "contentType is not application/json")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("handling request: %s", body)

	var obj runtime.Object
	var gvk *schema.GroupVersionKind
	obj, gvk, err = codecs.UniversalDeserializer().Decode(body, nil, nil)
	if err != nil {
		klog.ErrorS(err, "request body could not be decoded")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	switch *gvk {
	// TODO v1beta1 admissionReview
	case v1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1.AdmissionReview)
		if !ok {
			err = fmt.Errorf("expected v1.AdmissionReview but got: %T", obj)
			klog.ErrorS(err, "wrong object type")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		responseAdmissionReview := &v1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview

		klog.Infof("start writing response: %v", responseObj)

		var respBytes []byte
		respBytes, err = json.Marshal(responseObj)
		if err != nil {
			klog.ErrorS(err, "failed to marshal response object")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respBytes)
		if err != nil {
			klog.ErrorS(err, "failed to write response")
		}
	default:
		err = fmt.Errorf("unsupported group version kind: %v", gvk)
		klog.ErrorS(err, "unsupported group version kind")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	return
}

func startServer(ctx context.Context, tlsConfig *tls.Config, cw *CertWatcher, admitter *Admitter) error {
	go func() {
		klog.Info("Starting certificate watcher")
		if err := cw.Start(ctx); err != nil {
			klog.ErrorS(err, "failed to start certificate watcher")
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/persistentvolumeclaims", admitter.serverPVCRequest)
	srv := &http.Server{
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	// listener is always closed by srv.Serve
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		return err
	}
	return srv.Serve(listener)
}

func main(cmd *cobra.Command, args []string) {
	// Create new cert watcher
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()
	cw, err := NewCertWatcher(certFile, keyFile)
	if err != nil {
		klog.Fatalf("failed to initialize new cert watcher: %v", err)
	}
	tslConfig := &tls.Config{
		GetCertificate: cw.GetCertificate,
	}

	admitter, err := NewAdmitter()
	if err != nil {
		klog.Fatalf("failed to initialize new admitter: %v", err)
	}

	err = startServer(ctx, tslConfig, cw, admitter)
	if err != nil {
		klog.Fatalf("failed to start server: %v", err)
	}
}

package webhook

import (
	"context"
	"crypto/tls"
	"sync"

	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
)

// This file originated from github.com/kubernetes-sigs/controller-runtime/pkg/webhook/internal/certwatcher.
// We cannot import this package as it's an internal one. In addition, we cannot yet easily integrate
// with controller-runtime/pkg/webhook directly, as it would require extensive rework:
// https://github.com/kubernetes-csi/external-snapshotter/issues/422

// CertWatcher watches certificate and key files for changes.  When either file
// changes, it reads and parses both and calls an optional callback with the new
// certificate.
type CertWatcher struct {
	sync.Mutex

	currentCert *tls.Certificate
	watcher     *fsnotify.Watcher

	certPath string
	keyPath  string
}

// NewCertWatcher returns a new CertWatcher watching the given certificate and key.
func NewCertWatcher(certPath, keyPath string) (*CertWatcher, error) {
	var err error

	cw := &CertWatcher{
		certPath: certPath,
		keyPath:  keyPath,
	}

	// Initial read of certificate and key.
	if err := cw.ReadCertificate(); err != nil {
		return nil, err
	}

	cw.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return cw, nil
}

// GetCertificate fetches the currently loaded certificate, which may be nil.
func (cw *CertWatcher) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cw.Lock()
	defer cw.Unlock()
	return cw.currentCert, nil
}

// Start starts the watch on the certificate and key files.
func (cw *CertWatcher) Start(ctx context.Context) error {
	files := []string{cw.certPath, cw.keyPath}

	for _, f := range files {
		if err := cw.watcher.Add(f); err != nil {
			return err
		}
	}

	go cw.Watch()

	// Block until the context is done.
	<-ctx.Done()

	return cw.watcher.Close()
}

// Watch reads events from the watcher's channel and reacts to changes.
func (cw *CertWatcher) Watch() {
	for {
		select {
		case event, ok := <-cw.watcher.Events:
			// Channel is closed.
			if !ok {
				return
			}

			cw.handleEvent(event)

		case err, ok := <-cw.watcher.Errors:
			// Channel is closed.
			if !ok {
				return
			}

			klog.Error(err, "certificate watch error")
		}
	}
}

// ReadCertificate reads the certificate and key files from disk, parses them,
// and updates the current certificate on the watcher.  If a callback is set, it
// is invoked with the new certificate.
func (cw *CertWatcher) ReadCertificate() error {
	cert, err := tls.LoadX509KeyPair(cw.certPath, cw.keyPath)
	if err != nil {
		return err
	}

	cw.Lock()
	cw.currentCert = &cert
	cw.Unlock()

	klog.Info("Updated current TLS certificate")

	return nil
}

func (cw *CertWatcher) handleEvent(event fsnotify.Event) {
	// Only care about events which may modify the contents of the file.
	if !(isWrite(event) || isRemove(event) || isCreate(event)) {
		return
	}

	klog.V(1).Info("certificate event", "event", event)

	// If the file was removed, re-add the watch.
	if isRemove(event) {
		if err := cw.watcher.Add(event.Name); err != nil {
			klog.Error(err, "error re-watching file")
		}
	}

	if err := cw.ReadCertificate(); err != nil {
		klog.Error(err, "error re-reading certificate")
	}
}

func isWrite(event fsnotify.Event) bool {
	return event.Op&fsnotify.Write == fsnotify.Write
}

func isCreate(event fsnotify.Event) bool {
	return event.Op&fsnotify.Create == fsnotify.Create
}

func isRemove(event fsnotify.Event) bool {
	return event.Op&fsnotify.Remove == fsnotify.Remove
}

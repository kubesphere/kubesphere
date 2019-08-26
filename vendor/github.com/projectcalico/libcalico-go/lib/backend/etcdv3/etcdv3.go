// Copyright (c) 2016-2019 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcdv3

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crypto/tls"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/srv"
	"github.com/coreos/etcd/pkg/transport"
	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	"github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"
)

var (
	clientTimeout    = 10 * time.Second
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second
)

type etcdV3Client struct {
	etcdClient *clientv3.Client
}

func NewEtcdV3Client(config *apiconfig.EtcdConfig) (api.Client, error) {
	if config.EtcdEndpoints != "" && config.EtcdDiscoverySrv != "" {
		log.Warning("Multiple etcd endpoint discovery methods specified in etcdv3 API config")
		return nil, errors.New("multiple discovery or bootstrap options specified, use either \"etcdEndpoints\" or \"etcdDiscoverySrv\"")
	}

	// Split the endpoints into a location slice.
	etcdLocation := []string{}
	if config.EtcdEndpoints != "" {
		etcdLocation = strings.Split(config.EtcdEndpoints, ",")
	}

	if config.EtcdDiscoverySrv != "" {
		srvs, srvErr := srv.GetClient("etcd-client", config.EtcdDiscoverySrv)
		if srvErr != nil {
			return nil, fmt.Errorf("failed to discover etcd endpoints through SRV discovery: %v", srvErr)
		}
		etcdLocation = srvs.Endpoints
	}

	if len(etcdLocation) == 0 {
		log.Warning("No etcd endpoints specified in etcdv3 API config")
		return nil, errors.New("no etcd endpoints specified")
	}

	// Create the etcd client
	// If Etcd Certificate and Key are provided inline through command line agrument,
	// then the inline values take precedence over the ones in the config file.
	// All the three parametes, Certificate, key and CA certificate are to be provided inline for processing.
	var tls *tls.Config
	var err error

	haveInline := config.EtcdCert != "" || config.EtcdKey != "" || config.EtcdCACert != ""
	haveFiles := config.EtcdCertFile != "" || config.EtcdKeyFile != "" || config.EtcdCACertFile != ""

	if haveInline && haveFiles {
		return nil, fmt.Errorf("Cannot mix inline certificate-key and certificate / key files")
	}

	if haveInline {
		tlsInfo := &TlsInlineCertKey{
			CACert: config.EtcdCACert,
			Cert:   config.EtcdCert,
			Key:    config.EtcdKey,
		}
		tls, err = tlsInfo.ClientConfigInlineCertKey()
	} else {
		tlsInfo := &transport.TLSInfo{
			CAFile:   config.EtcdCACertFile,
			CertFile: config.EtcdCertFile,
			KeyFile:  config.EtcdKeyFile,
		}
		tls, err = tlsInfo.ClientConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("could not initialize etcdv3 client: %+v", err)
	}

	// Build the etcdv3 config.
	cfg := clientv3.Config{
		Endpoints:            etcdLocation,
		TLS:                  tls,
		DialTimeout:          clientTimeout,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
	}

	// Plumb through the username and password if both are configured.
	if config.EtcdUsername != "" && config.EtcdPassword != "" {
		cfg.Username = config.EtcdUsername
		cfg.Password = config.EtcdPassword
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &etcdV3Client{etcdClient: client}, nil
}

// Create an entry in the datastore.  If the entry already exists, this will return
// an ErrorResourceAlreadyExists error and the current entry.
func (c *etcdV3Client) Create(ctx context.Context, d *model.KVPair) (*model.KVPair, error) {
	logCxt := log.WithFields(log.Fields{"model-etcdKey": d.Key, "value": d.Value, "ttl": d.TTL, "rev": d.Revision})
	logCxt.Debug("Processing Create request")

	key, value, err := getKeyValueStrings(d)
	if err != nil {
		return nil, err
	}
	logCxt = logCxt.WithField("etcdv3-etcdKey", key)

	putOpts, err := c.getTTLOption(ctx, d)
	if err != nil {
		return nil, err
	}

	// Checking for 0 version of the etcdKey, which means it doesn't exists yet,
	// and if it does, get the current value.
	logCxt.Debug("Performing etcdv3 transaction for Create request")
	txnResp, err := c.etcdClient.Txn(ctx).If(
		clientv3.Compare(clientv3.Version(key), "=", 0),
	).Then(
		clientv3.OpPut(key, value, putOpts...),
	).Else(
		clientv3.OpGet(key),
	).Commit()
	if err != nil {
		logCxt.WithError(err).Warning("Create failed")
		return nil, cerrors.ErrorDatastoreError{Err: err}
	}

	if !txnResp.Succeeded {
		// The resource must already exist.  Extract the current newValue and
		// return that if possible.
		logCxt.Debug("Create transaction failed due to resource already existing")
		var existing *model.KVPair
		getResp := (*clientv3.GetResponse)(txnResp.Responses[0].GetResponseRange())
		if len(getResp.Kvs) != 0 {
			existing, _ = etcdToKVPair(d.Key, getResp.Kvs[0])
		}
		return existing, cerrors.ErrorResourceAlreadyExists{Identifier: d.Key}
	}

	v, err := model.ParseValue(d.Key, []byte(value))
	if err != nil {
		return nil, cerrors.ErrorPartialFailure{Err: fmt.Errorf("Unexpected error parsing stored datastore entry '%v': %+v", value, err)}
	}
	d.Value = v
	d.Revision = strconv.FormatInt(txnResp.Header.Revision, 10)

	return d, nil
}

// Update an entry in the datastore.  If the entry does not exist, this will return
// an ErrorResourceDoesNotExist error.  The ResourceVersion must be specified, and if
// incorrect will return a ErrorResourceUpdateConflict error and the current entry.
func (c *etcdV3Client) Update(ctx context.Context, d *model.KVPair) (*model.KVPair, error) {
	logCxt := log.WithFields(log.Fields{"model-etcdKey": d.Key, "value": d.Value, "ttl": d.TTL, "rev": d.Revision})
	logCxt.Debug("Processing Update request")
	key, value, err := getKeyValueStrings(d)
	if err != nil {
		return nil, err
	}
	logCxt = logCxt.WithField("etcdv3-etcdKey", key)

	opts, err := c.getTTLOption(ctx, d)
	if err != nil {
		return nil, err
	}

	// ResourceVersion must be set for an Update.
	rev, err := parseRevision(d.Revision)
	if err != nil {
		return nil, err
	}
	conds := []clientv3.Cmp{clientv3.Compare(clientv3.ModRevision(key), "=", rev)}

	logCxt.Debug("Performing etcdv3 transaction for Update request")
	txnResp, err := c.etcdClient.Txn(ctx).If(
		conds...,
	).Then(
		clientv3.OpPut(key, value, opts...),
	).Else(
		clientv3.OpGet(key),
	).Commit()

	if err != nil {
		logCxt.WithError(err).Warning("Update failed")
		return nil, cerrors.ErrorDatastoreError{Err: err}
	}

	// Etcd V3 does not return a error when compare condition fails we must verify the
	// response Succeeded field instead.  If the compare did not succeed then check for
	// a successful get to return either an UpdateConflict or a ResourceDoesNotExist error.
	if !txnResp.Succeeded {
		getResp := (*clientv3.GetResponse)(txnResp.Responses[0].GetResponseRange())
		if len(getResp.Kvs) == 0 {
			logCxt.Debug("Update transaction failed due to resource not existing")
			return nil, cerrors.ErrorResourceDoesNotExist{Identifier: d.Key}
		}

		logCxt.Debug("Update transaction failed due to resource update conflict")
		existing, _ := etcdToKVPair(d.Key, getResp.Kvs[0])
		return existing, cerrors.ErrorResourceUpdateConflict{Identifier: d.Key}
	}

	v, err := model.ParseValue(d.Key, []byte(value))
	cerrors.PanicIfErrored(err, "Unexpected error parsing stored datastore entry: %v", value)
	d.Value = v
	d.Revision = strconv.FormatInt(txnResp.Header.Revision, 10)

	return d, nil
}

//TODO Remove once we get rid of the v1 client.  Apply should no longer be supported
// at least in it's current guise.  Apply will need to be handled further up the stack
// by performing a Get/Create or Update to ensure we don't lose certain read-only Metadata.
// It's possible that we will just perform that processing in the clients (e.g. calicoctl),
// but that is to be decided.
func (c *etcdV3Client) Apply(ctx context.Context, d *model.KVPair) (*model.KVPair, error) {
	logCxt := log.WithFields(log.Fields{"etcdKey": d.Key, "value": d.Value, "ttl": d.TTL, "rev": d.Revision})
	logCxt.Debug("Processing Apply request")
	key, value, err := getKeyValueStrings(d)
	if err != nil {
		return nil, err
	}

	putOpts, err := c.getTTLOption(ctx, d)
	if err != nil {
		return nil, err
	}

	logCxt.Debug("Performing etcdv3 Put for Apply request")
	resp, err := c.etcdClient.Put(ctx, key, value, putOpts...)
	if err != nil {
		logCxt.WithError(err).Warning("Apply failed")
		return nil, cerrors.ErrorDatastoreError{Err: err}
	}

	v, err := model.ParseValue(d.Key, []byte(value))
	cerrors.PanicIfErrored(err, "Unexpected error parsing stored datastore entry: %v", value)
	d.Value = v
	d.Revision = strconv.FormatInt(resp.Header.Revision, 10)

	return d, nil
}

func (c *etcdV3Client) DeleteKVP(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	return c.Delete(ctx, kvp.Key, kvp.Revision)
}

// Delete an entry in the datastore.  This errors if the entry does not exists.
func (c *etcdV3Client) Delete(ctx context.Context, k model.Key, revision string) (*model.KVPair, error) {
	logCxt := log.WithFields(log.Fields{"model-etcdKey": k, "rev": revision})
	logCxt.Debug("Processing Delete request")
	key, err := model.KeyToDefaultDeletePath(k)
	if err != nil {
		return nil, err
	}
	logCxt = logCxt.WithField("etcdv3-etcdKey", key)

	conds := []clientv3.Cmp{}
	if len(revision) != 0 {
		rev, err := parseRevision(revision)
		if err != nil {
			return nil, err
		}
		conds = append(conds, clientv3.Compare(clientv3.ModRevision(key), "=", rev))
	}

	// Perform the delete transaction - note that this is an exact delete, not a prefix delete.
	logCxt.Debug("Performing etcdv3 transaction for Delete request")
	txnResp, err := c.etcdClient.Txn(ctx).If(
		conds...,
	).Then(
		clientv3.OpDelete(key, clientv3.WithPrevKV()),
	).Else(
		clientv3.OpGet(key),
	).Commit()
	if err != nil {
		logCxt.WithError(err).Warning("Delete failed")
		return nil, cerrors.ErrorDatastoreError{Err: err, Identifier: k}
	}

	// Transaction did not succeed - which means the ModifiedIndex check failed.  We can respond
	// with the latest settings.
	if !txnResp.Succeeded {
		logCxt.Debug("Delete transaction failed due to resource update conflict")

		getResp := txnResp.Responses[0].GetResponseRange()
		if len(getResp.Kvs) == 0 {
			logCxt.Debug("Delete transaction failed due to resource not existing")
			return nil, cerrors.ErrorResourceDoesNotExist{Identifier: k}
		}
		latestValue, err := etcdToKVPair(k, getResp.Kvs[0])
		if err != nil {
			return nil, err
		}
		return latestValue, cerrors.ErrorResourceUpdateConflict{Identifier: k}
	}

	// The delete response should have succeeded since the Get response did.
	delResp := txnResp.Responses[0].GetResponseDeleteRange()
	if delResp.Deleted == 0 {
		logCxt.Debug("Delete transaction failed due to resource not existing")
		return nil, cerrors.ErrorResourceDoesNotExist{Identifier: k}
	}

	// Parse the deleted value.  Don't propagate the error in this case since the
	// delete did succeed.
	previousValue, _ := etcdToKVPair(k, delResp.PrevKvs[0])
	return previousValue, nil
}

// Get an entry from the datastore.  This errors if the entry does not exist.
func (c *etcdV3Client) Get(ctx context.Context, k model.Key, revision string) (*model.KVPair, error) {
	logCxt := log.WithFields(log.Fields{"model-etcdKey": k, "rev": revision})
	logCxt.Debug("Processing Get request")

	key, err := model.KeyToDefaultPath(k)
	if err != nil {
		logCxt.Error("Unable to convert model.Key to an etcdv3 etcdKey")
		return nil, err
	}
	logCxt = logCxt.WithField("etcdv3-etcdKey", key)

	ops := []clientv3.OpOption{}
	if len(revision) != 0 {
		rev, err := parseRevision(revision)
		if err != nil {
			return nil, err
		}
		ops = append(ops, clientv3.WithRev(rev))
	}

	logCxt.Debug("Calling Get on etcdv3 client")
	resp, err := c.etcdClient.Get(ctx, key, ops...)
	if err != nil {
		logCxt.WithError(err).Debug("Error returned from etcdv3 client")
		return nil, cerrors.ErrorDatastoreError{Err: err}
	}
	if len(resp.Kvs) == 0 {
		logCxt.Debug("No results returned from etcdv3 client")
		return nil, cerrors.ErrorResourceDoesNotExist{Identifier: k}
	}

	return etcdToKVPair(k, resp.Kvs[0])
}

// List entries in the datastore.  This may return an empty list of there are
// no entries matching the request in the ListInterface.
func (c *etcdV3Client) List(ctx context.Context, l model.ListInterface, revision string) (*model.KVPairList, error) {
	logCxt := log.WithFields(log.Fields{"list-interface": l, "rev": revision})
	logCxt.Debug("Processing List request")

	// To list entries, we enumerate from the common root based on the supplied IDs, and then filter the results.
	key, ops := calculateListKeyAndOptions(logCxt, l)
	logCxt = logCxt.WithField("etcdv3-etcdKey", key)

	// We may also need to perform a get based on a particular revision.
	if len(revision) != 0 {
		rev, err := parseRevision(revision)
		if err != nil {
			return nil, err
		}
		ops = append(ops, clientv3.WithRev(rev))
	}

	logCxt.Debug("Calling Get on etcdv3 client")
	resp, err := c.etcdClient.Get(ctx, key, ops...)
	if err != nil {
		logCxt.WithError(err).Debug("Error returned from etcdv3 client")
		return nil, cerrors.ErrorDatastoreError{Err: err}
	}
	logCxt.WithField("numResults", len(resp.Kvs)).Debug("Processing response from etcdv3")

	// Filter/process the results.
	list := []*model.KVPair{}
	for _, p := range resp.Kvs {
		if kv := convertListResponse(p, l); kv != nil {
			list = append(list, kv)
		}
	}

	return &model.KVPairList{
		KVPairs:  list,
		Revision: strconv.FormatInt(resp.Header.Revision, 10),
	}, nil
}

func calculateListKeyAndOptions(logCxt *log.Entry, l model.ListInterface) (string, []clientv3.OpOption) {
	// -  If the final name segment of the name is itself a prefix, then just perform a prefix Get
	//    using the constructed key.
	// -  If the etcdKey is actually fully qualified, then perform an exact Get using the constructed
	//    key.
	// -  If the etcdKey is not fully qualified then it is a path prefix but the last segment is complete.
	//    Append a terminating "/" and perform a prefix Get.  The terminating / for a prefix Get ensures
	//    for a prefix of "/a" we only return "child entries" of "/a" such as "/a/x" and not siblings
	//    such as "/ab".
	key := model.ListOptionsToDefaultPathRoot(l)
	var ops []clientv3.OpOption
	if model.IsListOptionsLastSegmentPrefix(l) {
		// The last segment is a prefix, perform a prefix Get without adding a segment
		// delimiter.
		logCxt.Debug("List options is a name prefix, don't add a / to the path")
		ops = append(ops, clientv3.WithPrefix())
	} else if !model.ListOptionsIsFullyQualified(l) {
		// The etcdKey not a fully qualified etcdKey - it must be a prefix.
		logCxt.Debug("List options is a parent prefix, ensure path ends in /")
		if !strings.HasSuffix(key, "/") {
			logCxt.Debug("Adding / to path")
			key += "/"
		}
		ops = append(ops, clientv3.WithPrefix())
	}

	return key, ops
}

// EnsureInitialized makes sure that the etcd data is initialized for use by
// Calico.
func (c *etcdV3Client) EnsureInitialized() error {
	//TODO - still need to worry about ready flag.
	return nil
}

// Clean removes all of the Calico data from the datastore.
func (c *etcdV3Client) Clean() error {
	log.Warning("Cleaning etcdv3 datastore of all Calico data")
	_, err := c.etcdClient.Txn(context.Background()).If().Then(
		clientv3.OpDelete("/calico/", clientv3.WithPrefix()),
	).Commit()

	if err != nil {
		return cerrors.ErrorDatastoreError{Err: err}
	}
	return nil
}

// IsClean() returns true if there are no /calico/ prefixed entries in the
// datastore.  This is not part of the exposed API, but is public to allow
// direct consumers of the backend API to access this.
func (c *etcdV3Client) IsClean() (bool, error) {
	log.Debug("Calling Get on etcdv3 client")
	resp, err := c.etcdClient.Get(context.Background(), "/calico/", clientv3.WithPrefix())
	if err != nil {
		log.WithError(err).Debug("Error returned from etcdv3 client")
		return false, cerrors.ErrorDatastoreError{Err: err}
	}

	// The datastore is clean if no results were enumerated.
	return len(resp.Kvs) == 0, nil
}

// getTTLOption returns a OpOption slice containing a Lease granted for the TTL.
func (c *etcdV3Client) getTTLOption(ctx context.Context, d *model.KVPair) ([]clientv3.OpOption, error) {
	putOpts := []clientv3.OpOption{}

	if d.TTL != 0 {
		resp, err := c.etcdClient.Lease.Grant(ctx, int64(d.TTL.Seconds()))
		if err != nil {
			log.WithError(err).Error("Failed to grant a lease")
			return nil, cerrors.ErrorDatastoreError{Err: err}
		}

		putOpts = append(putOpts, clientv3.WithLease(resp.ID))
	}

	return putOpts, nil
}

// getKeyValueStrings returns the etcdv3 etcdKey and serialized value calculated from the
// KVPair.
func getKeyValueStrings(d *model.KVPair) (string, string, error) {
	logCxt := log.WithFields(log.Fields{"model-etcdKey": d.Key, "value": d.Value})
	key, err := model.KeyToDefaultPath(d.Key)
	if err != nil {
		logCxt.WithError(err).Error("Failed to convert model-etcdKey to etcdv3 etcdKey")
		return "", "", cerrors.ErrorDatastoreError{
			Err:        err,
			Identifier: d.Key,
		}
	}
	bytes, err := model.SerializeValue(d)
	if err != nil {
		logCxt.WithError(err).Error("Failed to serialize value")
		return "", "", cerrors.ErrorDatastoreError{
			Err:        err,
			Identifier: d.Key,
		}
	}

	return key, string(bytes), nil
}

// parseRevision parses the model.KVPair revision string and converts to the
// equivalent etcdv3 int64 value.
func parseRevision(revs string) (int64, error) {
	rev, err := strconv.ParseInt(revs, 10, 64)
	if err != nil {
		log.WithField("Revision", revs).Debug("Unable to parse Revision")
		return 0, cerrors.ErrorValidation{
			ErroredFields: []cerrors.ErroredField{
				{
					Name:  "ResourceVersion",
					Value: revs,
				},
			},
		}
	}
	return rev, nil
}

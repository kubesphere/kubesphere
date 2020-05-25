/*
Copyright 2019 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ldap

import (
	"crypto/tls"
	"log"
	"time"

	"github.com/go-ldap/ldap"
)

// PoolConn implements Client to override the Close() method
type PoolConn struct {
	Conn     ldap.Client
	c        *channelPool
	unusable bool
	closeAt  []uint16
}

func (p *PoolConn) Start() {
	p.Conn.Start()
}

func (p *PoolConn) StartTLS(config *tls.Config) error {
	// FIXME - check if already TLS and then ignore?
	return p.Conn.StartTLS(config)
}

// Close() puts the given connects back to the pool instead of closing it.
func (p *PoolConn) Close() {
	if p.unusable {
		log.Printf("Closing unusable connection")
		if p.Conn != nil {
			p.Conn.Close()
		}
		return
	}
	p.c.put(p.Conn)
}

func (p *PoolConn) SimpleBind(simpleBindRequest *ldap.SimpleBindRequest) (*ldap.SimpleBindResult, error) {
	return p.Conn.SimpleBind(simpleBindRequest)
}

func (p *PoolConn) Bind(username, password string) error {
	return p.Conn.Bind(username, password)
}

func (p *PoolConn) ModifyDN(modifyDNRequest *ldap.ModifyDNRequest) error {
	return p.Conn.ModifyDN(modifyDNRequest)
}

// MarkUnusable() marks the connection not usable any more, to let the pool close it
// instead of returning it to pool.
func (p *PoolConn) MarkUnusable() {
	p.unusable = true
}

func (p *PoolConn) autoClose(err error) {
	for _, code := range p.closeAt {
		if ldap.IsErrorWithCode(err, code) {
			p.MarkUnusable()
			return
		}
	}
}

func (p *PoolConn) SetTimeout(t time.Duration) {
	p.Conn.SetTimeout(t)
}

func (p *PoolConn) Add(addRequest *ldap.AddRequest) error {
	return p.Conn.Add(addRequest)
}

func (p *PoolConn) Del(delRequest *ldap.DelRequest) error {
	return p.Conn.Del(delRequest)
}

func (p *PoolConn) Modify(modifyRequest *ldap.ModifyRequest) error {
	return p.Conn.Modify(modifyRequest)
}

func (p *PoolConn) Compare(dn, attribute, value string) (bool, error) {
	return p.Conn.Compare(dn, attribute, value)
}

func (p *PoolConn) PasswordModify(passwordModifyRequest *ldap.PasswordModifyRequest) (*ldap.PasswordModifyResult, error) {
	return p.Conn.PasswordModify(passwordModifyRequest)
}

func (p *PoolConn) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	return p.Conn.Search(searchRequest)
}
func (p *PoolConn) SearchWithPaging(searchRequest *ldap.SearchRequest, pagingSize uint32) (*ldap.SearchResult, error) {
	return p.Conn.SearchWithPaging(searchRequest, pagingSize)
}

/*
Copyright 2020 The KubeSphere Authors.

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

package v1alpha3

import v1 "k8s.io/api/core/v1"

/**
We use a special type of secret as a credential for DevOps.
This file will not contain CRD, but the credential type constants and their fields.
*/
const (
	CredentialFinalizerName  = "finalizers.kubesphere.io/credential"
	DevOpsCredentialPrefix   = "credential.devops.kubesphere.io/"
	DevOpsCredentialDataHash = DevOpsCredentialPrefix + "datahash"
	// SecretTypeBasicAuth contains data needed for basic authentication.
	//
	// Required at least one of fields:
	// - Secret.Data["username"] - username used for authentication
	// - Secret.Data["password"] - password or token needed for authentication
	SecretTypeBasicAuth v1.SecretType = DevOpsCredentialPrefix + "basic-auth"
	// BasicAuthUsernameKey is the key of the username for SecretTypeBasicAuth secrets
	BasicAuthUsernameKey = "username"
	// BasicAuthPasswordKey is the key of the password or token for SecretTypeBasicAuth secrets
	BasicAuthPasswordKey = "password"

	// SecretTypeSSHAuth contains data needed for ssh authentication.
	//
	// Required at least one of fields:
	// - Secret.Data["username"] - username used for authentication
	// - Secret.Data["passphrase"] - passphrase needed for authentication
	// - Secret.Data["privatekey"] - privatekey needed for authentication
	SecretTypeSSHAuth v1.SecretType = DevOpsCredentialPrefix + "ssh-auth"
	// SSHAuthUsernameKey is the key of the username for SecretTypeSSHAuth secrets
	SSHAuthUsernameKey = "username"
	// SSHAuthPrivateKey is the key of the passphrase for SecretTypeSSHAuth secrets
	SSHAuthPassphraseKey = "passphrase"
	// SSHAuthPrivateKey is the key of the privatekey for SecretTypeSSHAuth secrets
	SSHAuthPrivateKey = "private_key"

	// SecretTypeSecretText contains data.
	//
	// Required at least one of fields:
	// - Secret.Data["secret"] - secret
	SecretTypeSecretText v1.SecretType = DevOpsCredentialPrefix + "secret-text"
	// SecretTextSecretKey is the key of the secret for SecretTypeSecretText secrets
	SecretTextSecretKey = "secret"

	// SecretTypeKubeConfig contains data.
	//
	// Required at least one of fields:
	// - Secret.Data["secret"] - secret
	SecretTypeKubeConfig v1.SecretType = DevOpsCredentialPrefix + "kubeconfig"
	// KubeConfigSecretKey is the key of the secret for SecretTypeKubeConfig secrets
	KubeConfigSecretKey = "content"
	//	CredentialAutoSyncAnnoKey is used to indicate whether the secret is automatically synchronized to devops.
	//	In the old version, the credential is stored in jenkins and cannot be obtained.
	//	This field is set to ensure that the secret is not overwritten by a nil value.
	CredentialAutoSyncAnnoKey   = DevOpsCredentialPrefix + "autosync"
	CredentialSyncStatusAnnoKey = DevOpsCredentialPrefix + "syncstatus"
	CredentialSyncTimeAnnoKey   = DevOpsCredentialPrefix + "synctime"
	CredentialSyncMsgAnnoKey    = DevOpsCredentialPrefix + "syncmsg"
)

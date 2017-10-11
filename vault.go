package cfutil

import (
	"errors"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
	vault "github.com/hashicorp/vault/api"
)

type VaultClient struct {
	vault.Client
	Endpoint           string
	RoleID             string
	SecretID           string
	ServiceSecretPath  string
	ServiceTransitPath string
	SpaceSecretPath    string
	OrgSecretPath      string
	Secret             *vault.Secret
}

func (v *VaultClient) Login() (err error) {
	path := "auth/approle/login"
	options := map[string]interface{}{
		"role_id":   v.RoleID,
		"secret_id": v.SecretID,
	}
	v.Secret, err = v.Logical().Write(path, options)
	return err
}

func NewVaultClient(serviceName string) (*VaultClient, error) {
	appEnv, _ := Current()
	service := &cfenv.Service{}
	err := errors.New("")
	if serviceName != "" {
		service, err = serviceByName(appEnv, serviceName)
	} else {
		service, err = serviceByTag(appEnv, "Vault")
	}
	if err != nil {
		return nil, err
	}

	var vaultClient VaultClient

	if str, ok := service.Credentials["role_id"].(string); ok {
		vaultClient.RoleID = str
	}
	if str, ok := service.Credentials["secret_id"].(string); ok {
		vaultClient.SecretID = str
	}
	if str, ok := service.Credentials["org_secret_path"].(string); ok {
		vaultClient.OrgSecretPath = str
	}
	if str, ok := service.Credentials["service_secret_path"].(string); ok {
		vaultClient.ServiceSecretPath = str
	}
	if str, ok := service.Credentials["endpoint"].(string); ok {
		vaultClient.Endpoint = str
	}
	if str, ok := service.Credentials["space_secret_path"].(string); ok {
		vaultClient.SpaceSecretPath = str
	}
	if str, ok := service.Credentials["service_transit_path"].(string); ok {
		vaultClient.ServiceTransitPath = str
	}

	client, err := vault.NewClient(&vault.Config{
		Address: vaultClient.Endpoint,
	})
	if err != nil {
		return nil, err
	}
	vaultClient.Client = *client

	return &vaultClient, vaultClient.Login()
}

package cfutil

import (
	"errors"
	"regexp"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
	vault "github.com/hashicorp/vault/api"
)

var v1Regex = regexp.MustCompile(`/v1/`)

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
	v.SetToken(v.Secret.Auth.ClientToken)
	return err
}

func (v *VaultClient) ReadSpaceString(path string) (string, error) {
	return v.ReadString(v.SpaceSecretPath, path)
}

func (v *VaultClient) ReadOrgString(path string) (string, error) {
	return v.ReadString(v.OrgSecretPath, path)
}

func (v *VaultClient) ReadString(prefix, path string) (string, error) {
	err := v.Login()
	if err != nil {
		return "", err
	}
	secret, err := v.Logical().Read(prefix + "/" + path)
	if err != nil {
		return "", err
	}
	str, ok := secret.Data["value"].(string)
	if !ok {
		return "", errors.New("Missing value string")
	}
	return str, nil
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
		vaultClient.OrgSecretPath = v1Regex.ReplaceAllString(str, "")
	}
	if str, ok := service.Credentials["service_secret_path"].(string); ok {
		vaultClient.ServiceSecretPath = v1Regex.ReplaceAllString(str, "")
	}
	if str, ok := service.Credentials["endpoint"].(string); ok {
		vaultClient.Endpoint = str
	}
	if str, ok := service.Credentials["space_secret_path"].(string); ok {
		vaultClient.SpaceSecretPath = v1Regex.ReplaceAllString(str, "")
	}
	if str, ok := service.Credentials["service_transit_path"].(string); ok {
		vaultClient.ServiceTransitPath = v1Regex.ReplaceAllString(str, "")
	}

	client, err := vault.NewClient(&vault.Config{
		Address: vaultClient.Endpoint,
	})
	if err != nil {
		return nil, err
	}
	vaultClient.Client = *client
	err = vaultClient.Login()
	if err != nil {
		return nil, err
	}
	return &vaultClient, vaultClient.Login()
}

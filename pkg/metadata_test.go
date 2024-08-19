package pkg

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/luci/go-render/render"
)

var metadataMarshalData = map[string]marshalData{
	"nil": {
		Data:   []byte(`null`),
		Object: nil,
	},
	"empty": {
		Data:   []byte(`{}`),
		Object: Metadata{},
	},
	"only federation": {
		Data: []byte(`{"federation_entity":{"contacts":["contact@federation.endpoint"],"federation_fetch_endpoint":"https://federation.endpoint/fetch","federation_list_endpoint":"https://federation.endpoint/list","federation_resolve_endpoint":"https://federation.endpoint/resolve","federation_trust_mark_status_endpoint":"https://federation.endpoint/trustmarks","organization_name":"Test Organization"}}`),
		Object: Metadata{
			FederationEntity: &FederationEntityMetadata{
				FederationFetchEndpoint:           "https://federation.endpoint/fetch",
				FederationListEndpoint:            "https://federation.endpoint/list",
				FederationResolveEndpoint:         "https://federation.endpoint/resolve",
				FederationTrustMarkStatusEndpoint: "https://federation.endpoint/trustmarks",
				OrganizationName:                  "Test Organization",
				Contacts:                          []string{"contact@federation.endpoint"},
			},
		},
	},
	"federation + op": {
		Data: []byte(`{"openid_provider":{"authorization_endpoint":"https://op.example.com/authorization","client_registration_types_supported":["automatic"],"code_challenge_methods_supported":["S256"],"grant_types_supported":["authorization_code","refresh_token"],"id_token_signed_response_alg_values_supported":["ES256","ES512"],"introspection_endpoint":"https://op.example.com/introspect","issuer":"https://op.example.com","jwks_uri":"https://op.example.com/jwks","organization_name":"Test OP Org","request_signed_response_alg_values_supported":["ES256","ES512"],"request_uri_parameter_supported":false,"response_types_supported":["code"],"revocation_endpoint":"https://op.example.com/revoke","scopes_supported":["openid","profile","email","offline_access"],"subject_types_supported":null,"token_endpoint":"https://op.example.com/token","userinfo_endpoint":"https://op.example.com/userinfo","userinfo_signed_response_alg_values_supported":["ES256","ES512"]},"federation_entity":{"contacts":["contact@op.example.com"],"organization_name":"Test OP Org"}}`),
		Object: Metadata{
			OpenIDProvider: &OpenIDProviderMetadata{
				Issuer:                "https://op.example.com",
				AuthorizationEndpoint: "https://op.example.com/authorization",
				TokenEndpoint:         "https://op.example.com/token",
				UserinfoEndpoint:      "https://op.example.com/userinfo",
				ScopesSupported: []string{
					"openid",
					"profile",
					"email",
					"offline_access",
				},
				ResponseTypesSupported: []string{"code"},
				GrantTypesSupported: []string{
					"authorization_code",
					"refresh_token",
				},
				IDTokenSignedResponseAlgValuesSupported: []string{
					"ES256",
					"ES512",
				},
				UserinfoSignedResponseAlgValuesSupported: []string{
					"ES256",
					"ES512",
				},
				RequestSignedResponseAlgValuesSupported: []string{
					"ES256",
					"ES512",
				},
				RevocationEndpoint:               "https://op.example.com/revoke",
				IntrospectionEndpoint:            "https://op.example.com/introspect",
				CodeChallengeMethodsSupported:    []string{"S256"},
				ClientRegistrationTypesSupported: []string{"automatic"},
				OrganizationName:                 "Test OP Org",
				JWKSURI:                          "https://op.example.com/jwks",
			},
			FederationEntity: &FederationEntityMetadata{
				OrganizationName: "Test OP Org",
				Contacts:         []string{"contact@op.example.com"},
			},
		},
	},
	"federation + rp": {
		Data: []byte(`{"openid_relying_party":{"application_type":"web","client_name":"test rp","client_registration_types":["automatic"],"client_uri":"https://rp.example.org","contacts":["contact@rp.example.org"],"grant_types":["authorization_code"],"jwks_uri":"https://rp.example.org/jwks","organization_name":"test rp org","policy_uri":"https://rp.example.org/policy","redirect_uris":["https://rp.example.org/redirect"],"response_types":["code"],"scope":"openid profile email","tos_uri":"https://rp.example.org/tos"},"federation_entity":{"contacts":["contact@rp.example.org"],"organization_name":"test rp org"}}`),
		Object: Metadata{
			RelyingParty: &OpenIDRelyingPartyMetadata{
				Scope:                   "openid profile email",
				RedirectURIS:            []string{"https://rp.example.org/redirect"},
				ResponseTypes:           []string{"code"},
				GrantTypes:              []string{"authorization_code"},
				ApplicationType:         "web",
				Contacts:                []string{"contact@rp.example.org"},
				ClientName:              "test rp",
				ClientURI:               "https://rp.example.org",
				PolicyURI:               "https://rp.example.org/policy",
				TOSURI:                  "https://rp.example.org/tos",
				JWKSURI:                 "https://rp.example.org/jwks",
				OrganizationName:        "test rp org",
				ClientRegistrationTypes: []string{"automatic"},
			},
			FederationEntity: &FederationEntityMetadata{
				OrganizationName: "test rp org",
				Contacts:         []string{"contact@rp.example.org"},
			},
		},
	},
	"op extra fields": {
		Data: []byte(`{"openid_provider":{"authorization_endpoint":"","client_registration_types_supported":null,"foo":"bar","issuer":"","request_uri_parameter_supported":false,"response_types_supported":null,"slice":["two","values"],"subject_types_supported":null,"token_endpoint":""}}`),
		Object: Metadata{
			OpenIDProvider: &OpenIDProviderMetadata{
				Extra: map[string]interface{}{
					"foo": "bar",
					"slice": []any{
						"two",
						"values",
					},
				},
			},
		},
	},
	"rp extra fields": {
		Data: []byte(`{"openid_relying_party":{"client_registration_types":null,"foo":"bar","slice":["two","values"]}}`),
		Object: Metadata{
			RelyingParty: &OpenIDRelyingPartyMetadata{
				Extra: map[string]interface{}{
					"foo": "bar",
					"slice": []any{
						"two",
						"values",
					},
				},
			},
		},
	},
	"as extra fields": {
		Data: []byte(`{"oauth_authorization_server":{"authorization_endpoint":"","client_registration_types_supported":null,"foo":"bar","issuer":"","request_uri_parameter_supported":false,"response_types_supported":null,"slice":["two","values"],"subject_types_supported":null,"token_endpoint":""}}`),
		Object: Metadata{
			OAuthAuthorizationServer: &OAuthAuthorizationServerMetadata{
				Extra: map[string]interface{}{
					"foo": "bar",
					"slice": []any{
						"two",
						"values",
					},
				},
			},
		},
	},
	"client extra fields": {
		Data: []byte(`{"oauth_client":{"client_registration_types":null,"foo":"bar","slice":["two","values"]}}`),
		Object: Metadata{
			OAuthClient: &OAuthClientMetadata{
				Extra: map[string]interface{}{
					"foo": "bar",
					"slice": []any{
						"two",
						"values",
					},
				},
			},
		},
	},
	"pr extra fields": {
		Data: []byte(`{"oauth_resource":{"foo":"bar","resource_encryption_alg_values_supported":null,"resource_encryption_enc_values_supported":null,"slice":["two","values"]}}`),
		Object: Metadata{
			OAuthProtectedResource: &OAuthProtectedResourceMetadata{
				Extra: map[string]interface{}{
					"foo": "bar",
					"slice": []any{
						"two",
						"values",
					},
				},
			},
		},
	},
	"federation extra fields": {
		Data: []byte(`{"federation_entity":{"foo":"bar","slice":["two","values"]}}`),
		Object: Metadata{
			FederationEntity: &FederationEntityMetadata{
				Extra: map[string]interface{}{
					"foo": "bar",
					"slice": []any{
						"two",
						"values",
					},
				},
			},
		},
	},
}

func TestMetadata_MarshalJSON(t *testing.T) {
	for name, test := range metadataMarshalData {
		t.Run(
			name, func(t *testing.T) {
				j, err := json.Marshal(test.Object)
				if err != nil {
					t.Error(err)
				}
				if !bytes.Equal(j, test.Data) {
					t.Errorf("Marshal result not as expected.\nExpected: %s\n     Got: %s", test.Data, j)
				}
			},
		)
	}
}

func TestMetadata_UnmarshalJSON(t *testing.T) {
	for name, test := range metadataMarshalData {
		t.Run(
			name, func(t *testing.T) {
				var result Metadata
				err := json.Unmarshal(test.Data, &result)
				if err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(test.Object, result) {
					if name == "nil" {
						// the nil test is only for marshalling not unmarshalling into nil
						return
					}
					t.Errorf(
						"Unmarshal result not as expected.\nExpected: %s\n     Got: %s", render.Render(test.Object),
						render.Render(result),
					)
				}
			},
		)
	}
}

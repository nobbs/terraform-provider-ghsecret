// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/lithammer/dedent"
)

// Ensure GHSecretProvider satisfies various provider interfaces.
var _ provider.Provider = &GHSecretProvider{}
var _ provider.ProviderWithFunctions = &GHSecretProvider{}

// GHSecretProvider defines the provider implementation.
type GHSecretProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GHSecretProviderModel describes the provider data model.
type GHSecretProviderModel struct {
}

func (p *GHSecretProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ghsecret"
	resp.Version = p.version
}

func (p *GHSecretProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: strings.TrimSpace(dedent.Dedent(`
			The GitHub Secrets provider allows you to encrypt plain text to be stored as an pre-encrypted
			GitHub secret via the [github](https://registry.terraform.io/providers/hashicorp/github) provider.
			This way, no plain text secrets need to be stored in the Terraform state.

			The data is encrypted using the [libsodium sealed box encryption scheme](https://libsodium.gitbook.io/doc/public-key_cryptography/sealed_boxes)
			and the [public key of the GitHub repository](https://docs.github.com/en/rest/actions/secrets?apiVersion=2022-11-28#get-a-repository-public-key)
			where the secret will be stored. The encrypted data is then base64 encoded and returned as a
			string ready to be used in the GitHub provider.

			The encryption is done using the [anonymous sealed box encryption](https://pkg.go.dev/golang.org/x/crypto/nacl/box#SealAnonymous)
			provided by the Go implementation of libsodium. Part of this encryption scheme is the generation
			of a ephemeral key pair that is used to encrypt the data with the public key of the GitHub
			repository. The ephemeral public key is included in the encrypted data and is used by the
			GitHub repository to decrypt the data. This way, the data can only be decrypted by the GitHub
			repository that has the corresponding private key.

			**Caution:** the ephemeral key pair requires a random number generator for secure key generation.
			This provider uses a random number generator seeded with data derived from the hash of the cleartext
			data to ensure that the ephemeral key pair is deterministic for a given cleartext. This is done,
			since Terraform expects constant results for a given input and the encryption otherwise would
			result in different outputs for the same input on each run. This is most likely not to be
			considered secure in a **strict cryptographic sense**, but still an improvement over ending up with
			plain text secrets in the Terraform state.
		`)),
		Attributes: map[string]schema.Attribute{},
	}
}

func (p *GHSecretProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data GHSecretProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *GHSecretProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *GHSecretProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *GHSecretProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewEncryptFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GHSecretProvider{
			version: version,
		}
	}
}

// Copyright (c) Alexej Disterhoft <alexej@disterhoft.de>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/lithammer/dedent"
	"golang.org/x/crypto/nacl/box"
)

func helperEncryptFunctionConfig(plaintext, pkB64 string) string {
	return strings.TrimSpace(dedent.Dedent(fmt.Sprintf(`
		output "encrypted" {
			value = provider::ghsecret::encrypt("%s", "%s")
			sensitive = true
		}
	`, plaintext, pkB64),
	),
	)
}

func TestEncryptFunction(t *testing.T) {
	// Generate a new public/private key pair.
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: helperEncryptFunctionConfig("hello world", base64.StdEncoding.EncodeToString(publicKey[:])),
				Check: resource.ComposeTestCheckFunc(
					testDecrypt("hello world", *publicKey, *privateKey),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"encrypted",
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"encrypted": knownvalue.NotNull(),
							"sha256":    knownvalue.NotNull(),
						}),
					),
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testDecrypt(plaintext string, publicKey, privateKey [32]byte) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := "encrypted"

		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		// get the encrypted value
		valueMap, ok := rs.Value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Output '%s': expected map, got %#v", name, rs)
		}

		enc, ok := valueMap["encrypted"].(string)
		if !ok {
			return fmt.Errorf("Output '%s': expected string, got %#v", name, rs)
		}

		// base64 decode the encrypted value
		encrypted, err := base64.StdEncoding.DecodeString(enc)
		if err != nil {
			return fmt.Errorf("failed to decode base64: %v", err)
		}

		// decrypt the encrypted value
		dec, ok := box.OpenAnonymous(nil, encrypted, &publicKey, &privateKey)
		if !ok {
			return fmt.Errorf("failed to decrypt")
		}

		decrypted := string(dec)

		// compare the decrypted value with the expected value
		if decrypted != plaintext {
			return fmt.Errorf(
				"Output '%s': expected %#v, got %#v",
				name,
				plaintext,
				rs)
		}

		return nil
	}
}

// Copyright (c) Alexej Disterhoft <alexej@disterhoft.de>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lithammer/dedent"
	"golang.org/x/crypto/nacl/box"
)

var encryptReturnAttrTypes = map[string]attr.Type{
	"encrypted": types.StringType,
	"sha256":    types.StringType,
}

// Ensure that encryptFunction implements the Function interface.
var _ function.Function = &encryptFunction{}

type encryptFunction struct{}

func NewEncryptFunction() function.Function {
	return &encryptFunction{}
}

func (f *encryptFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "encrypt"
}

func (f *encryptFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Encrypt a plain text to be stored as a GitHub secret.",
		MarkdownDescription: strings.TrimSpace(dedent.Dedent(`
			Encrypts a plain text to be stored as a GitHub secret. The data is encrypted using the
			libsodium sealed box encryption scheme and the public key of a GitHub repository. The
			encrypted data is then base64 encoded and returned as a string.
			`)),
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "plaintext",
				MarkdownDescription: "The plain text to encrypt.",
			},
			function.StringParameter{
				Name:                "public_key",
				MarkdownDescription: "The base64 encoded public key of the GitHub repository.",
			},
		},

		Return: function.ObjectReturn{
			AttributeTypes: encryptReturnAttrTypes,
		},
	}
}

func (f *encryptFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var plaintext, publicKey string

	resp.Error = req.Arguments.Get(ctx, &plaintext, &publicKey)
	if resp.Error != nil {
		return
	}

	// Encrypt the plaintext using the public key
	encrypted, err := encrypt(plaintext, publicKey)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("failed to encrypt data: %v", err))
		return
	}

	encryptedB64 := base64.StdEncoding.EncodeToString(encrypted)

	// Calculate the SHA256 hash of the plaintext
	hashBytes := sha256.Sum256([]byte(plaintext))
	hash := base64.StdEncoding.EncodeToString(hashBytes[:])

	result, diags := types.ObjectValue(
		encryptReturnAttrTypes, map[string]attr.Value{
			"encrypted": types.StringValue(encryptedB64),
			"sha256":    types.StringValue(hash),
		},
	)

	resp.Error = function.FuncErrorFromDiags(ctx, diags)
	if resp.Error != nil {
		return
	}

	resp.Error = resp.Result.Set(ctx, &result)
}

// encrypt encrypts the plaintext using the public key. The public key must be base64 encoded.
func encrypt(plaintext, pkB64 string) ([]byte, error) {
	pkBytes, err := base64.StdEncoding.DecodeString(pkB64)
	if err != nil {
		return nil, err
	}

	var pkBytes32 [32]byte
	copiedLen := copy(pkBytes32[:], pkBytes)
	if copiedLen == 0 {
		return nil, fmt.Errorf("could not convert publicKey to bytes")
	}

	plaintextBytes := []byte(plaintext)
	var encryptedBytes []byte

	cipherText, err := box.SealAnonymous(encryptedBytes, plaintextBytes, &pkBytes32, nil)
	if err != nil {
		return nil, err
	}

	return cipherText, nil
}

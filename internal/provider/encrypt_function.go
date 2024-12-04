// Copyright (c) Alexej Disterhoft <alexej@disterhoft.de>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lithammer/dedent"
	"golang.org/x/crypto/nacl/box"
)

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

		Return: function.StringReturn{},
	}
}

func (f *encryptFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var plaintext, publicKey string

	resp.Error = req.Arguments.Get(ctx, &plaintext, &publicKey)
	if resp.Error != nil {
		return
	}

	encrypted, err := encrypt(plaintext, publicKey)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("failed to encrypt data: %v", err))
		return
	}

	// Base64 encode the encrypted data.
	encryptedB64 := base64.StdEncoding.EncodeToString(encrypted)
	result := types.StringValue(encryptedB64)

	resp.Error = resp.Result.Set(ctx, &result)
}

// encrypt encrypts the given secret using the provided base64-encoded public key.
// It returns the encrypted secret as a byte slice or an error if the encryption fails.
func encrypt(secret, pkB64 string) ([]byte, error) {
	decodedPubKey, err := base64.StdEncoding.DecodeString(pkB64)
	if err != nil {
		return nil, err
	}

	var pubKey [32]byte
	copy(pubKey[:], decodedPubKey)

	secretBytes := []byte(secret)

	cipherText, err := box.SealAnonymous(nil, secretBytes, &pubKey, deterministicRand(secret))
	if err != nil {
		return nil, err
	}

	return cipherText, nil
}

// deterministicRand returns a new rand.Rand seeded with a deterministic value derived from the
// input string.
//
// Is it cryptograhpically secure? Probably not. Is this a problem? Not really, the rand.Rand is
// only used to generate an ephemeral key pair for the encryption, of which the private key is
// discarded immediately after one use. The public key is passed along with the encrypted data and
// is not sensitive.
func deterministicRand(s string) *rand.Rand {
	seed := sha256.Sum256([]byte(s))
	seedInt := int64(binary.LittleEndian.Uint64(seed[:8]))
	return rand.New(rand.NewSource(seedInt))
}

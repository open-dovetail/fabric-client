/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPIs
*/

package signcert

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grantae/certinfo"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	pvmsp "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// return value at path c1.c2.c3 from yaml file, does not handle arrays
func execYamlPath(node interface{}, path string) interface{} {
	tokens := strings.Split(path, ".")
	result := node
	var ok bool
	for _, name := range tokens {
		result, ok = yamlChildNode(result, name)
		if !ok {
			return nil
		}
	}
	return result
}

func yamlChildNode(parent interface{}, name string) (interface{}, bool) {
	data, ok := parent.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}
	c, ok := data[name]
	return c, ok
}

// UserCertificate returns certificate string of a specified user@org
func UserCertificate(user string) string {
	userTokens := strings.Split(user, "@")
	u := userTokens[0]
	org := ""
	if len(userTokens) > 1 {
		org = userTokens[1]
	}
	var data map[interface{}]interface{}
	yaml.Unmarshal(NetworkConfig, &data)
	cryptoPath := execYamlPath(data, "client.cryptoconfig.path").(string)
	if len(org) == 0 {
		// use network client org if user org is not specified
		org = execYamlPath(data, "client.organization").(string)
	}

	// find cert file for specified user and org
	var certStore core.KVStore
	var mspid interface{}
	var err error
	for k, v := range data["organizations"].(map[interface{}]interface{}) {
		if k.(string) == org {
			mspid, _ = yamlChildNode(v, "mspid")
			if pathTemplate, ok := yamlChildNode(v, "cryptoPath"); ok {
				if !filepath.IsAbs(pathTemplate.(string)) {
					pathTemplate = filepath.Join(cryptoPath, pathTemplate.(string))
				}
				certStore, err = msp.NewFileCertStore(Subst(pathTemplate.(string)))
			}
			break
		}
	}
	if err != nil || certStore == nil {
		logger.Debugf("cannot find crypto path for org %s", org)
		return ""
	}

	// read the cert file
	cert, err := certStore.Load(&pvmsp.IdentityIdentifier{
		ID:    u,
		MSPID: mspid.(string),
	})
	if err != nil {
		logger.Debugf("cannot read cert file of %s@%s", u, org)
		return ""
	}

	// extract cert info from the certificate
	block, rest := pem.Decode(cert.([]byte))
	if block == nil || len(rest) > 0 {
		logger.Debugf("failed to decode pem cert file of %s@%s", u, org)
		return ""
	}
	blkBytes, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		logger.Debugf("failed to parse x509 certificate of %s@%s", u, org)
		return ""
	}

	// print out cert info
	certText, err := certinfo.CertificateText(blkBytes)
	if err != nil {
		logger.Debugf("failed to print cert info of %s@%s", u, org)
		return ""
	}
	return certText
}

// ReadFile returns content of a specified file
func ReadFile(filePath string) ([]byte, error) {
	f, err := os.Open(Subst(filePath))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open file: %s", filePath)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read file stat: %s", filePath)
	}
	s := fi.Size()
	cBytes := make([]byte, s)
	n, err := f.Read(cBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read file: %s", filePath)
	}
	if n == 0 {
		fmt.Printf("file %s is empty\n", filePath)
	}
	return cBytes, err
}

// Subst replaces instances of '${VARNAME}' (eg ${GOPATH}) with the variable.
// Variables names that are not set by the SDK are replaced with the environment variable.
func Subst(path string) string {
	const (
		sepPrefix = "${"
		sepSuffix = "}"
	)

	splits := strings.Split(path, sepPrefix)

	var buffer bytes.Buffer

	// first split precedes the first sepPrefix so should always be written
	buffer.WriteString(splits[0]) // nolint: gas

	for _, s := range splits[1:] {
		subst, rest := substVar(s, sepPrefix, sepSuffix)
		buffer.WriteString(subst) // nolint: gas
		buffer.WriteString(rest)  // nolint: gas
	}

	return buffer.String()
}

// substVar searches for an instance of a variables name and replaces them with their value.
// The first return value is substituted portion of the string or noMatch if no replacement occurred.
// The second return value is the unconsumed portion of s.
func substVar(s string, noMatch string, sep string) (string, string) {
	endPos := strings.Index(s, sep)
	if endPos == -1 {
		return noMatch, s
	}

	v, ok := os.LookupEnv(s[:endPos])
	if !ok {
		return noMatch, s
	}

	return v, s[endPos+1:]
}

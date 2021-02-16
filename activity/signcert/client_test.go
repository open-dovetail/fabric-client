/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package signcert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test requires to start the fabric test-network using
//    network.sh up createChannel
//    network.sh deployCC
// and set network config files as in setup of activity_test.go

func TestUserCert(t *testing.T) {
	cert := UserCertificate("User1")
	logger.Infof("user cert: %s\n", cert)
	assert.Contains(t, cert, "CN=User1@org1.example.com", "cert info should contain User1 as cn")
}

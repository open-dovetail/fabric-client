module github.com/open-dovetail/fabric-client/activity/signcert

go 1.14

replace github.com/project-flogo/flow => github.com/yxuco/flow v1.1.1

replace github.com/project-flogo/core => github.com/yxuco/core v1.2.2

replace go.uber.org/multierr => go.uber.org/multierr v1.6.0

replace github.com/grantae/certinfo => github.com/yxuco/certinfo v0.0.1

require (
	github.com/grantae/certinfo v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-sdk-go v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/project-flogo/core v1.2.0
	github.com/stretchr/testify v1.6.1
	go.uber.org/multierr v1.6.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

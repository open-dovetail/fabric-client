module github.com/open-dovetail/fabric-client/activity/request

go 1.14

replace github.com/project-flogo/flow => github.com/yxuco/flow v1.1.1

replace github.com/project-flogo/core => github.com/yxuco/core v1.2.2

replace go.uber.org/multierr => go.uber.org/multierr v1.6.0

require (
	github.com/hyperledger/fabric-sdk-go v1.0.0-rc1
	github.com/pkg/errors v0.9.1
	github.com/project-flogo/core v1.2.0
	github.com/stretchr/testify v1.6.1
	github.com/xeipuuv/gojsonschema v1.1.0
	go.uber.org/multierr v1.6.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

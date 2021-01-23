module github.com/open-dovetail/fabric-client/plugin

go 1.14

replace github.com/project-flogo/cli => github.com/yxuco/cli v0.10.1-0.20201211003232-196e588c1452

require (
	github.com/open-dovetail/fabric-chaincode/plugin v0.1.4
	github.com/open-dovetail/fabric-client/activity/request v0.0.1
	github.com/pkg/errors v0.9.1
	github.com/project-flogo/cli v0.10.0
	github.com/project-flogo/core v1.2.0
	github.com/project-flogo/flow v1.2.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	github.com/xeipuuv/gojsonschema v1.1.0
)

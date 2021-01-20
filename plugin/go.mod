module github.com/open-dovetail/fabric-client/plugin

go 1.14

replace github.com/project-flogo/cli => github.com/yxuco/cli v0.10.1-0.20201211003232-196e588c1452

require (
	github.com/open-dovetail/fabric-client/activity/request v0.0.1
	github.com/project-flogo/cli v0.10.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
)

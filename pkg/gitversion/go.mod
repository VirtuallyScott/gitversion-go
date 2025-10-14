module github.com/VirtuallyScott/gitversion-go/pkg/gitversion

go 1.21

require (
	github.com/VirtuallyScott/gitversion-go/pkg/config v0.0.0
	github.com/VirtuallyScott/gitversion-go/pkg/semver v0.0.0
)

replace (
	github.com/VirtuallyScott/gitversion-go/pkg/config => ../config
	github.com/VirtuallyScott/gitversion-go/pkg/semver => ../semver
)
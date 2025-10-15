module github.com/VirtuallyScott/gitversion-go

go 1.21

require (
	github.com/VirtuallyScott/gitversion-go/internal/version v0.0.0
	github.com/VirtuallyScott/gitversion-go/pkg/config v0.0.0
	github.com/VirtuallyScott/gitversion-go/pkg/gitversion v0.0.0
	github.com/VirtuallyScott/gitversion-go/pkg/semver v0.0.0
	gopkg.in/yaml.v3 v3.0.1
)

replace (
	github.com/VirtuallyScott/gitversion-go/internal/version => ./internal/version
	github.com/VirtuallyScott/gitversion-go/pkg/config => ./pkg/config
	github.com/VirtuallyScott/gitversion-go/pkg/gitversion => ./pkg/gitversion
	github.com/VirtuallyScott/gitversion-go/pkg/semver => ./pkg/semver
)

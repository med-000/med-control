module github.com/med-000/med-control/backend

go 1.25

require (
	github.com/med-000/med-control/infra v0.0.0
	github.com/med-000/med-control/shared v0.0.0
)

replace github.com/med-000/med-control/infra => ../infra

replace github.com/med-000/med-control/shared => ../shared

module github.com/med-000/overview/backend

go 1.25

require (
	github.com/med-000/overview/infra v0.0.0
	github.com/med-000/overview/shared v0.0.0
)

replace github.com/med-000/overview/infra => ../infra

replace github.com/med-000/overview/shared => ../shared

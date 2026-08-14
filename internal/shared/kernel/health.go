package kernel

type IHealth interface {
	Health(ctx Ctx) error
}

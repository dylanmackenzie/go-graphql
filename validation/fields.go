package validation

func report(ctx *execution.Context, s string, v ...interface{}) {
	ctx.addErrors(log.Errorf(s, v...))
}

func checkField(f *ast.Field, desc schema.AbstractDescriptor, ctx *execution.Context) {
	name := Field.Name

	field, ok := desc.Field(name)
	if !ok {
		report("Field '%s' not present on '%s'", name, desc.Name())
	}
}

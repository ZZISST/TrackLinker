package ports

type FindOption interface {
	Apply(*Option)
}

type Option struct {
	Join     string
	Query    []Query
	Order    any
	Group    string
	Having   []Query
	Offset   int
	Limit    int
	Preloads []string
}

type optionFn func(*Option)

func (f optionFn) Apply(opt *Option) {
	f(opt)
}

func WithQuery(query ...Query) FindOption {
	return optionFn(func(opt *Option) {
		opt.Query = query
	})
}

func WithOffset(offset int) FindOption {
	return optionFn(func(opt *Option) {
		opt.Offset = offset
	})
}

func WithLimit(limit int) FindOption {
	return optionFn(func(opt *Option) {
		opt.Limit = limit
	})
}

func WithOrder(order interface{}) FindOption {
	return optionFn(func(opt *Option) {
		opt.Order = order
	})
}

func WithGroup(group string) FindOption {
	return optionFn(func(opt *Option) {
		opt.Group = group
	})
}

func WithHaving(having ...Query) FindOption {
	return optionFn(func(opt *Option) {
		opt.Having = having
	})
}

func WithPreload(preloads []string) FindOption {
	return optionFn(func(opt *Option) {
		opt.Preloads = preloads
	})
}

func WithJoin(join string) FindOption {
	return optionFn(func(opt *Option) {
		opt.Join = join
	})
}

package komi

// poolWork is an internal function type to set pool's work performer.
type poolWork[I, O any] func(p *Pool[I, O])

// noValue is an internal type used to indicate whether a pool is producing
// outputs or not. So if output's type is `noValue`, then no outputs are made.
type noValue any

// WorkSimple should be used to set work with no outputs nor errors.
func WorkSimple[I any](work func(I)) poolWork[I, noValue] {
	return func(p *Pool[I, noValue]) {
		p.workSimple = work
		p.workPerformer = p.performWorkSimple
	}
}

// WorkSimpleWithErrors should be used to set work with no outputs but with errors.
func WorkSimpleWithErrors[I any](work func(I) error) poolWork[I, noValue] {
	return func(p *Pool[I, noValue]) {
		p.workSimpleWithErrors = work
		p.workPerformer = p.performWorkSimpleWithErrors
	}
}

// Work should be used to set work with outputs but no errors.
func Work[I, O any](work func(I) O) poolWork[I, O] {
	return func(p *Pool[I, O]) {
		p.workRegular = work
		p.workPerformer = p.performWorkRegular
	}
}

// WorkWithErrors should be used to set work with both outputs and errors.
func WorkWithErrors[I, O any](work func(I) (O, error)) poolWork[I, O] {
	return func(p *Pool[I, O]) {
		p.workRegularWithErrors = work
		p.workPerformer = p.performWorkWithErrors
	}
}

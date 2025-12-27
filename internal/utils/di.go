package utils

import "go.uber.org/dig"

func Resolve[T any](container *dig.Container) (T, error) {
	var dependency T
	err := container.Invoke(func(d T) {
		dependency = d
	})
	return dependency, err
}

func MustResolve[T any](container *dig.Container) T {
	dependency, err := Resolve[T](container)
	if err != nil {
		panic(err)
	}
	return dependency
}

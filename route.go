package jvmao

type routeChache struct {
	chache map[string]string
}

func newRouteChache() *routeChache {
	return &routeChache{chache: map[string]string{}}
}

func (rc *routeChache) SetRoute(name, pattern string) {
	rc.chache[name] = pattern
}
